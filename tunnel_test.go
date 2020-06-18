package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseDomain(t *testing.T) {
	tests := []struct {
		topDomain string
		domain    string
		fails     bool
		output    msgFragment
	}{
		{
			topDomain: "tunnel.example.com.",
			domain:    "2jkhm3.592.0.jf2ca2ltebqxiidxn5zgwidfozsxe6lxnbsxezjmebthk3tdoruw63tjnztsa43.nn5xxi2dmpeqgc5baoruw2zltfqqgc5ban52gqzlseb2gs3lfomqgs3ramzuxi4.zamfxgiidtorqxe5dtfyqes5bamjzgkylunbsxglbanf2ca2dfmf2hglbanf2ca.zlborzs4icjoqqhg2djorzsaylomq.tunnel.example.com.",
			output: msgFragment{
				id:        "2jkhm3",
				totalSize: 592,
				offset:    0,
				data:      "jf2ca2ltebqxiidxn5zgwidfozsxe6lxnbsxezjmebthk3tdoruw63tjnztsa43nn5xxi2dmpeqgc5baoruw2zltfqqgc5ban52gqzlseb2gs3lfomqgs3ramzuxi4zamfxgiidtorqxe5dtfyqes5bamjzgkylunbsxglbanf2ca2dfmf2hglbanf2cazlborzs4icjoqqhg2djorzsaylomq",
			},
			fails: false,
		},
		{
			topDomain: "tunnel.example.com.",
			domain:    "2jkhm3.FAIL.0.jf2ca2ltebqxiidxn5zgwidfozsxe6lxnbsxezjmebthk3tdoruw63tjnztsa43.nn5xxi2dmpeqgc5baoruw2zltfqqgc5ban52gqzlseb2gs3lfomqgs3ramzuxi4.zamfxgiidtorqxe5dtfyqes5bamjzgkylunbsxglbanf2ca2dfmf2hglbanf2ca.zlborzs4icjoqqhg2djorzsaylomq.tunnel.example.com.",
			fails:     true,
		},
		{
			topDomain: "tunnel.example.com.",
			domain:    "2jkhm3.592.FAIL.jf2ca2ltebqxiidxn5zgwidfozsxe6lxnbsxezjmebthk3tdoruw63tjnztsa43.nn5xxi2dmpeqgc5baoruw2zltfqqgc5ban52gqzlseb2gs3lfomqgs3ramzuxi4.zamfxgiidtorqxe5dtfyqes5bamjzgkylunbsxglbanf2ca2dfmf2hglbanf2ca.zlborzs4icjoqqhg2djorzsaylomq.tunnel.example.com.",
			fails:     true,
		},
		{
			topDomain: "tunnel.example.com.",
			domain:    "2jkhm3.592.0.tunnel.example.com.",
			fails:     true,
		},
	}
	for _, test := range tests {
		got, err := parseDomain(test.topDomain, test.domain)
		if test.fails {
			require.NotNil(t, err)
		} else {
			require.Nil(t, err)
		}
		require.Equal(t, test.output, got)
	}
}
