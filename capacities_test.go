package packngo

import (
	"fmt"
	"testing"
)

func TestAccCheckCapacity(t *testing.T) {
	skipUnlessAcceptanceTestsAllowed(t)
	c, stopRecord := setup(t)
	defer stopRecord()

	input := &CapacityInput{
		[]ServerInfo{
			{
				Facility: "ams1",
				Plan:     "baremetal_0",
				Quantity: 1},
		},
	}

	cap, _, err := c.CapacityService.Check(input)
	if err != nil {
		t.Fatal(err)
	}

	for _, s := range cap.Servers {
		if !s.Available {
			t.Fatal(fmt.Errorf("capacity of %d servers should have been available", input.Servers[0].Quantity))
			break
		}
	}

	list, _, err := c.CapacityService.List()
	if err != nil {
		t.Fatal(err)
	}

	for k, v := range *list {
		if v["baremetal_2a2"].Level == "unavailable" {
			input.Servers[0].Plan = "baremetal_2a2"
			input.Servers[0].Facility = k
			break
		}
	}

	cap, _, err = c.CapacityService.Check(input)
	if err != nil {
		t.Fatal(err)
	}

	for _, s := range cap.Servers {
		if s.Available {
			t.Fatal(fmt.Errorf("capacity of %d servers should not have been available", input.Servers[0].Quantity))
			break
		}
	}
}
