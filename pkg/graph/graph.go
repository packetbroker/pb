// SPDX-FileCopyrightText: Copyright 2021 The Things Industries B.V.
// SPDX-License-Identifier: Apache-2.0

// Package graph writes Packet Broker reports in Graphviz format.
package graph

import (
	"context"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/emicklei/dot"
	reportingpb "go.packetbroker.org/api/reporting"
	packetbroker "go.packetbroker.org/api/v3"
)

const (
	rankDir            = "LR"
	nodeSep            = 0.2
	rankSep            = 1.0
	fontName           = "Helvetica"
	styleNoSuccess     = "dashed"
	fontSizeN          = "10pt"
	fontSizeHighlightN = "14pt"
	colorNoSuccess     = "gray50"
	fontSizeSuccessE   = "10pt"
	fontSizeNoSuccessE = "8pt"
	fontSizeNoSuccessN = "8pt"
	widthScaleE        = 2 // pt to scale the width of an edge
)

func tenantNode(
	parent *dot.Graph,
	tenantID packetbroker.TenantID,
	networks map[packetbroker.TenantID]*packetbroker.NetworkOrTenant,
	highlight bool,
) dot.Node {
	cluster := parent.Subgraph(tenantID.NetID.String(), dot.ClusterOption{})
	cluster.Attr("fontname", fontName)
	if host, ok := networks[packetbroker.TenantID{NetID: tenantID.NetID}]; ok {
		cluster = cluster.Label(fmt.Sprintf("%s (%s)", host.GetNetwork().GetName(), tenantID.NetID))
	}
	node, ok := cluster.FindNodeById(tenantID.String())
	if !ok {
		node = cluster.Node(tenantID.String())
		switch nwk, ok := networks[tenantID]; {
		case ok && nwk.GetNetwork().GetName() != "":
			node.Label(nwk.GetNetwork().GetName()).Box()
		case ok && nwk.GetTenant().GetName() != "":
			node.Label(nwk.GetTenant().GetName())
		case tenantID.ID == "":
			node.Label(tenantID.NetID.String()).Box()
		default:
			node.Label(tenantID.ID)
		}
		if highlight {
			node.Attrs(
				"fontsize", fontSizeHighlightN,
				"style", "filled",
			)
		} else {
			node.Attr("fontsize", fontSizeN)
		}
	}
	return node
}

// itoaShort formats the given integer to a short string by using a thousands unit: K, M, B or T.
// One fractional digit is preserved.
func itoaShort(v uint64) string {
	units := []string{"", "K", "M", "B", "T"}
	p := int(math.Floor(math.Log10(float64(v)))) / 3
	if maxUnit := len(units) - 1; p > maxUnit {
		p = maxUnit
	}
	if p > 0 {
		return fmt.Sprintf("%.1f%s", math.Floor(float64(v)/math.Pow10(3*p-1))/10, units[p])
	}
	return strconv.FormatUint(v, 10)
}

// WriteRoutedMessages writes the records of routed messages in Graphviz format.
// The networks map provides names that are shown instead of NetID and Tenant ID.
// The optional highlight argument indicates the identifier of the node to highlight.
func WriteRoutedMessages(
	out io.Writer,
	records []*reportingpb.RoutedMessagesRecord,
	networks map[packetbroker.TenantID]*packetbroker.NetworkOrTenant,
	highlight *packetbroker.TenantID,
) error {
	dotGraph := dot.NewGraph(dot.Directed)
	dotGraph.Attrs(
		"rankdir", rankDir,
		"nodesep", dot.Literal(strconv.FormatFloat(nodeSep, 'f', 1, 32)),
		"ranksep", dot.Literal(strconv.FormatFloat(rankSep, 'f', 1, 32)),
	)
	dotGraph.NodeInitializer(func(n dot.Node) {
		n.Attr("fontname", fontName)
	})
	dotGraph.EdgeInitializer(func(e dot.Edge) {
		e.Attr("fontname", fontName)
	})

	// Record the Forwarder and Home Network nodes and edges to scale the edge width based on relative score.
	type nodes struct {
		forwarderID,
		homeNetworkID packetbroker.TenantID
	}
	type nodesScore struct {
		forwarderNode,
		homeNetworkNode dot.Node
		edge  dot.Edge
		score float64
	}
	nodeScores := map[nodes]nodesScore{}
	maxScore := 0.0

	// Add the nodes and edges.
	for _, record := range records {
		var (
			uplink, downlink = record.GetUplink(), record.GetDownlink()
			forwarderID      = packetbroker.ForwarderTenantID(record)
			forwarderNode    = tenantNode(dotGraph, forwarderID, networks, highlight != nil && forwarderID == *highlight)
			homeNetworkID    = packetbroker.HomeNetworkTenantID(record)
			homeNetworkNode  = tenantNode(dotGraph, homeNetworkID, networks, highlight != nil && homeNetworkID == *highlight)
			totalUp          = uplink.GetDataMessagesRoutedCount() + uplink.GetJoinRequestsRoutedCount()
			totalDown        = downlink.GetDataMessagesRoutedCount() + downlink.GetJoinAcceptsRoutedCount()
			successUp        = uplink.GetDataMessagesProcessedSuccessCount() + uplink.GetJoinRequestsProcessedSuccessCount()
			successDown      = downlink.GetDataMessagesProcessedSuccessCount() +
				downlink.GetJoinAcceptsProcessedSuccessCount()
			label []string
			score float64
			attrs []any
		)
		switch {
		case successUp == 0 && successDown == 0:
			if totalUp > 0 {
				label = append(label, fmt.Sprintf("&uarr; %s", itoaShort(totalUp)))
			}
			if totalDown > 0 {
				label = append(label, fmt.Sprintf("&darr; %s", itoaShort(totalDown)))
			}
			attrs = append(attrs,
				"style", styleNoSuccess,
				"color", colorNoSuccess,
				"fontcolor", colorNoSuccess,
				"fontsize", fontSizeNoSuccessE,
			)
		default:
			if totalUp > 0 {
				label = append(label, fmt.Sprintf("&uarr; %s (%.0f%%)",
					itoaShort(successUp), float64(successUp)/float64(totalUp)*100.0),
				)
			}
			if totalDown > 0 {
				label = append(label, fmt.Sprintf("&darr; %s (%.0f%%)",
					itoaShort(successDown), float64(successDown)/float64(totalDown)*100.0),
				)
			}
			score = float64(successUp + successDown)
			attrs = append(attrs,
				"fontsize", fontSizeSuccessE,
			)
		}
		edge := dotGraph.Edge(forwarderNode, homeNetworkNode).
			Label(dot.HTML(strings.Join(label, "<br/>"))).
			Attr("weight", dot.Literal(strconv.FormatFloat(score, 'f', 0, 32)))
		edge.Attrs(attrs...)
		nodeScores[nodes{forwarderID, homeNetworkID}] = nodesScore{forwarderNode, homeNetworkNode, edge, score}
		if score > maxScore {
			maxScore = score
		}
	}

	// Apply the scores to the edge widths.
	// If there's a node with a zero score, apply the style indicating no successful messages routed.
	for ids, nodeScore := range nodeScores {
		if nodeScore.score > 0 {
			scaledScore := widthScaleE * math.Log2(nodeScore.score) / math.Log2(maxScore)
			nodeScore.edge.Attr("penwidth", dot.Literal(strconv.FormatFloat(scaledScore, 'f', 2, 32)))
		} else if highlight != nil {
			for _, entry := range []struct {
				packetbroker.TenantID
				dot.Node
			}{
				{ids.forwarderID, nodeScore.forwarderNode},
				{ids.homeNetworkID, nodeScore.homeNetworkNode},
			} {
				if entry.TenantID != *highlight {
					entry.Node.
						Attr("style", styleNoSuccess).
						Attr("fontsize", fontSizeNoSuccessN).
						Attr("color", colorNoSuccess)
				}
			}
		}
	}

	dotGraph.Write(out)
	if _, err := out.Write(nil); err != nil {
		return fmt.Errorf("graph: write: %w", err)
	}
	return nil
}

// RunDot executes the dot executable to convert the Graphviz input to an output format.
func RunDot(ctx context.Context, input io.Reader, output io.Writer, format string) error {
	//nolint:gosec // running dot with the given format is the purpose of this function
	cmd := exec.CommandContext(ctx, "dot", fmt.Sprintf("-T%s", format))
	cmd.Stdin, cmd.Stdout, cmd.Stderr = input, output, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run dot: %w", err)
	}
	return nil
}
