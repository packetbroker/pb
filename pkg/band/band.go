// Copyright © 2020 The Things Industries B.V.

package band

import packetbroker "go.packetbroker.org/api/v1beta3"

// DataRate is a data rate.
type DataRate interface{}

// DataRates defines the data rate by index.
type DataRates map[int]DataRate

// FindIndex returns the index by matching the given data rate.
func (d DataRates) FindIndex(dr DataRate) (int, bool) {
	switch t := dr.(type) {
	case LoRaDataRate:
		for i, dr := range d {
			ldr, ok := dr.(LoRaDataRate)
			if ok && t.SpreadingFactor == ldr.SpreadingFactor && t.Bandwidth == ldr.Bandwidth &&
				(t.Direction == ldr.Direction || ldr.Direction == Both) {
				return i, true
			}
		}
	case FSKDataRate:
		for i, dr := range d {
			fdr, ok := dr.(FSKDataRate)
			if ok && t.BitRate == fdr.BitRate {
				return i, true
			}
		}
	}
	return 0, false
}

// Direction indicates the transmission direction.
type Direction int

// LoRaDataRate is a LoRa data rate.
type LoRaDataRate struct {
	SpreadingFactor int
	Bandwidth       int
	Direction       Direction
}

// FSKDataRate is an FSK data rate.
type FSKDataRate struct {
	BitRate int
}

const (
	// Uplink is from end-device to network.
	Uplink Direction = iota
	// Downlink is from network to end device.
	Downlink
	// Both in uplink and downlink.
	Both
)

// RegionDataRates contains LoRaWAN data rates defined in LoRaWAN Regional Parmaters 1.1 revision B.
var RegionDataRates = map[packetbroker.Region]DataRates{
	packetbroker.Region_EU_863_870: DataRates{
		0: LoRaDataRate{12, 125000, Both},
		1: LoRaDataRate{11, 125000, Both},
		2: LoRaDataRate{10, 125000, Both},
		3: LoRaDataRate{9, 125000, Both},
		4: LoRaDataRate{8, 125000, Both},
		5: LoRaDataRate{7, 125000, Both},
		6: LoRaDataRate{7, 250000, Both},
		7: FSKDataRate{50000},
	},
	packetbroker.Region_US_902_928: DataRates{
		0:  LoRaDataRate{10, 125000, Uplink},
		1:  LoRaDataRate{9, 125000, Uplink},
		2:  LoRaDataRate{8, 125000, Uplink},
		3:  LoRaDataRate{7, 125000, Uplink},
		4:  LoRaDataRate{8, 500000, Uplink},
		8:  LoRaDataRate{12, 500000, Downlink},
		9:  LoRaDataRate{11, 500000, Downlink},
		10: LoRaDataRate{10, 500000, Downlink},
		11: LoRaDataRate{9, 500000, Downlink},
		12: LoRaDataRate{8, 500000, Downlink},
		13: LoRaDataRate{7, 500000, Downlink},
	},
	packetbroker.Region_CN_779_787: DataRates{
		0: LoRaDataRate{12, 125000, Both},
		1: LoRaDataRate{11, 125000, Both},
		2: LoRaDataRate{10, 125000, Both},
		3: LoRaDataRate{9, 125000, Both},
		4: LoRaDataRate{8, 125000, Both},
		5: LoRaDataRate{7, 125000, Both},
		6: LoRaDataRate{7, 250000, Both},
		7: FSKDataRate{50000},
	},
	packetbroker.Region_EU_433: DataRates{
		0: LoRaDataRate{12, 125000, Both},
		1: LoRaDataRate{11, 125000, Both},
		2: LoRaDataRate{10, 125000, Both},
		3: LoRaDataRate{9, 125000, Both},
		4: LoRaDataRate{8, 125000, Both},
		5: LoRaDataRate{7, 125000, Both},
		6: LoRaDataRate{7, 250000, Both},
		7: FSKDataRate{50000},
	},
	packetbroker.Region_AU_915_928: DataRates{
		0:  LoRaDataRate{12, 125000, Uplink},
		1:  LoRaDataRate{11, 125000, Uplink},
		2:  LoRaDataRate{10, 125000, Uplink},
		3:  LoRaDataRate{9, 125000, Uplink},
		4:  LoRaDataRate{8, 125000, Uplink},
		5:  LoRaDataRate{7, 125000, Uplink},
		6:  LoRaDataRate{8, 500000, Uplink},
		8:  LoRaDataRate{12, 500000, Downlink},
		9:  LoRaDataRate{11, 500000, Downlink},
		10: LoRaDataRate{10, 500000, Downlink},
		11: LoRaDataRate{9, 500000, Downlink},
		12: LoRaDataRate{8, 500000, Downlink},
		13: LoRaDataRate{7, 500000, Downlink},
	},
	packetbroker.Region_CN_470_510: DataRates{
		0: LoRaDataRate{12, 125000, Both},
		1: LoRaDataRate{11, 125000, Both},
		2: LoRaDataRate{10, 125000, Both},
		3: LoRaDataRate{9, 125000, Both},
		4: LoRaDataRate{8, 125000, Both},
		5: LoRaDataRate{7, 125000, Both},
	},
	packetbroker.Region_AS_923: DataRates{
		0: LoRaDataRate{12, 125000, Both},
		1: LoRaDataRate{11, 125000, Both},
		2: LoRaDataRate{10, 125000, Both},
		3: LoRaDataRate{9, 125000, Both},
		4: LoRaDataRate{8, 125000, Both},
		5: LoRaDataRate{7, 125000, Both},
		6: LoRaDataRate{7, 250000, Both},
		7: FSKDataRate{50000},
	},
	packetbroker.Region_KR_920_923: DataRates{
		0: LoRaDataRate{12, 125000, Both},
		1: LoRaDataRate{11, 125000, Both},
		2: LoRaDataRate{10, 125000, Both},
		3: LoRaDataRate{9, 125000, Both},
		4: LoRaDataRate{8, 125000, Both},
		5: LoRaDataRate{7, 125000, Both},
	},
	packetbroker.Region_IN_865_867: DataRates{
		0: LoRaDataRate{12, 125000, Both},
		1: LoRaDataRate{11, 125000, Both},
		2: LoRaDataRate{10, 125000, Both},
		3: LoRaDataRate{9, 125000, Both},
		4: LoRaDataRate{8, 125000, Both},
		5: LoRaDataRate{7, 125000, Both},
		7: FSKDataRate{50000},
	},
	packetbroker.Region_RU_864_870: DataRates{
		0: LoRaDataRate{12, 125000, Both},
		1: LoRaDataRate{11, 125000, Both},
		2: LoRaDataRate{10, 125000, Both},
		3: LoRaDataRate{9, 125000, Both},
		4: LoRaDataRate{8, 125000, Both},
		5: LoRaDataRate{7, 125000, Both},
		6: LoRaDataRate{7, 250000, Both},
		7: FSKDataRate{50000},
	},
}
