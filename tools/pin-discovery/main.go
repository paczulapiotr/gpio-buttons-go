package main

import (
	"fmt"
	"log"
	"sort"

	gpiocdev "github.com/warthog618/go-gpiocdev"
)

func main() {
	fmt.Println("Discovering available GPIO chips/lines...")
	fmt.Println("=====================================")

	chipNames := gpiocdev.Chips()
	if len(chipNames) == 0 {
		fmt.Println("❌ No gpiochips found!")
		fmt.Println("\nTroubleshooting:")
		fmt.Println("1. Run as root (sudo)")
		fmt.Println("2. Kernel must have GPIO chardev (CONFIG_GPIO_CDEV)")
		fmt.Println("3. Check: ls /dev/gpiochip*")
		return
	}

	// Flatten lines into a list of (chip, offset, name)
	type lineInfo struct {
		chip   string
		offset int
		name   string
	}
	var lines []lineInfo
	for _, cname := range chipNames {
		c, err := gpiocdev.NewChip(cname)
		if err != nil {
			log.Printf("Skipping %s: %v", cname, err)
			continue
		}
		n := c.Lines()
		for off := 0; off < n; off++ {
			li, err := c.LineInfo(off)
			if err != nil {
				continue
			}
			lines = append(lines, lineInfo{chip: cname, offset: off, name: li.Name})
		}
		_ = c.Close()
	}
	sort.Slice(lines, func(i, j int) bool {
		if lines[i].chip == lines[j].chip {
			return lines[i].offset < lines[j].offset
		}
		return lines[i].chip < lines[j].chip
	})

	fmt.Println("Chip:Offset               | Bias Pull-Up Support | Name")
	fmt.Println("--------------------------|----------------------|------")

	supported := 0
	checked := 0
	for _, ln := range lines {
		checked++
		// Try to request the line as input with bias pull-up.
		// Many lines may be busy or not support bias; ignore busy errors gracefully.
		pullSupport := "❌ NO"
		l, err := gpiocdev.RequestLine(ln.chip, ln.offset, gpiocdev.AsInput, gpiocdev.WithPullUp, gpiocdev.WithConsumer("pin-discovery"))
		if err == nil {
			pullSupport = "✅ YES"
			supported++
			_ = l.Close()
		}

		label := fmt.Sprintf("%s:%d", ln.chip, ln.offset)
		fmt.Printf("%-26s | %-20s | %s\n", label, pullSupport, ln.name)
	}

	fmt.Println("\n=====================================")
	fmt.Printf("✅ Lines with pull-up support: %d (out of %d checked)\n\n", supported, checked)

	if supported == 0 {
		fmt.Println("⚠️  No lines accepted bias=pull-up via userspace request.")
		fmt.Println("This can mean: kernel < v5.5, driver doesn't expose bias via gpiocdev, or line is reserved.")
		fmt.Println("Options: use external pull resistors, or set bias in device tree overlays.")
	} else {
		fmt.Println("Example usage with this library:")
		fmt.Println("  PinName: \"gpiochip0:23\"  // chip:line format")
	}
}
