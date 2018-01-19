// client
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"text/tabwriter"

	"github.com/digitalocean/godo"
)

const (
	prompt = "droplet> "
)

func clientHandler(in io.ReadCloser, out io.Writer, n *sync.WaitGroup) {
	defer n.Done()
	prompt := func(w io.Writer) { fmt.Fprint(w, prompt) }
	input := bufio.NewScanner(in)

	for prompt(out); input.Scan(); prompt(out) {
		args := strings.Split(input.Text(), " ")
		switch args[0] {
		case "quit", "q":
			close(requests)
			close(abort)
			for range replies {
				// drain it!
			}
			return
		default:
			requests <- &request{cmd: args[0], args: args[1:]}
			reply, ok := <-replies
			if !ok {
				fmt.Fprintln(out, "Server disconnected")
				return
			}
			if reply.err != nil {
				fmt.Fprintln(out, reply.err)
				break
			}
			printReplyData(out, reply)
		}
	}
}

func printReplyData(out io.Writer, r *reply) {
	switch r.dataType {
	case account:
		var a godo.Account
		if err := json.Unmarshal(r.data, &a); err != nil {
			fmt.Fprintln(out, err)
			break
		}
		printAccounts(out, a)
	case droplet:
		var d godo.Droplet
		if err := json.Unmarshal(r.data, &d); err != nil {
			fmt.Fprintln(out, err)
			break
		}
		printDroplets(out, d)
	case droplets:
		var dd []godo.Droplet
		if err := json.Unmarshal(r.data, &dd); err != nil {
			fmt.Fprintln(out, err)
			break
		}
		printDroplets(out, dd...)
	case keys:
		var kk []godo.Key
		if err := json.Unmarshal(r.data, &kk); err != nil {
			fmt.Fprintln(out, err)
			break
		}
		printKeys(out, kk...)
	case httpStatus:
		var status string
		if err := json.Unmarshal(r.data, &status); err != nil {
			fmt.Fprintln(out, err)
		}
		fmt.Fprintln(out, status)
	case invalid:
		fmt.Fprintf(out, "%s\n", r.data)
	default:
		fmt.Fprintf(out, "%s\n", r)
	}
}

func printAccounts(out io.Writer, accounts ...godo.Account) {
	const format = "%v\t%v\t%v\t%v\t%v\n"
	tw := new(tabwriter.Writer).Init(out, 0, 8, 2, ' ', 0)
	fmt.Fprintf(tw, format, "E-mail", "Status", "Droplet Limit", "Floating IP Limit", "UUID")
	fmt.Fprintf(tw, format, "------", "------", "-------------", "-----------------", "----")
	for _, a := range accounts {
		fmt.Fprintf(tw, format, a.Email, a.Status, a.DropletLimit,
			a.FloatingIPLimit, a.UUID)
	}
	tw.Flush()
}

func printDroplets(out io.Writer, droplets ...godo.Droplet) {
	const format = "%v\t%v\t%v\t%v\t%v\n"
	tw := new(tabwriter.Writer).Init(out, 0, 8, 2, ' ', 0)
	fmt.Fprintf(tw, format, "Identifier", "Droplet Name", "IPv4(public)", "IPv4(private)", "IPv6")
	fmt.Fprintf(tw, format, "----------", "------------", "------------", "-------------", "----")
	for _, d := range droplets {
		public := publicIPv4(&d)
		private := privateIPv4(&d)
		ipv6 := publicIPv6(&d)
		fmt.Fprintf(tw, format, d.ID, d.Name, public, private, ipv6)
	}
	tw.Flush()
}

func publicIPv4(d *godo.Droplet) string {
	public, err := d.PublicIPv4()
	if err != nil {
		return "*"
	}
	for _, ipv4 := range d.Networks.V4 {
		if ipv4.Type != "public" {
			continue
		}
		mask := net.ParseIP(ipv4.Netmask)
		if mask == nil {
			continue
		}
		mask = mask.To4()
		if mask == nil {
			continue
		}
		plen, _ := net.IPMask(mask).Size()
		if plen == 0 {
			continue
		}
		return fmt.Sprintf("%s/%d", public, plen)
	}
	return "*"
}

func privateIPv4(d *godo.Droplet) string {
	private, err := d.PrivateIPv4()
	if err != nil {
		return "*"
	}
	for _, ipv4 := range d.Networks.V4 {
		if ipv4.Type != "private" {
			continue
		}
		mask := net.ParseIP(ipv4.Netmask)
		if mask == nil {
			continue
		}
		mask = mask.To4()
		if mask == nil {
			continue
		}
		plen, _ := net.IPMask(mask).Size()
		if plen == 0 {
			continue
		}
		return fmt.Sprintf("%s/%d", private, plen)
	}
	return "*"
}

func publicIPv6(d *godo.Droplet) string {
	ipv6, err := d.PublicIPv6()
	if err != nil {
		return "*"
	}
	ip := net.ParseIP(ipv6)
	if ip == nil {
		return "*"
	}
	ip = ip.To16()
	if ip == nil {
		return "*"
	}
	for _, v6 := range d.Networks.V6 {
		if v6.Type != "public" {
			return "*"
		}
		return fmt.Sprintf("%s/%d", ip, v6.Netmask)
	}
	return "*"
}

func printKeys(out io.Writer, keys ...godo.Key) {
	const format = "%v\t%v\t%v\t%v\n"
	tw := new(tabwriter.Writer).Init(out, 0, 8, 2, ' ', 0)
	fmt.Fprintf(tw, format, "ID", "Name", "FingerPrint", "PublicKey")
	fmt.Fprintf(tw, format, "--", "----", "-----------", "---------")
	for _, k := range keys {
		fmt.Fprintf(tw, format, k.ID, k.Name, k.Fingerprint, k.PublicKey)
	}
	tw.Flush()
}
