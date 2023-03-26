package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	c "github.com/lightninglabs/lndclient"
	invpkg "github.com/lightningnetwork/lnd/invoices"
	"github.com/lightningnetwork/lnd/lnrpc"
)

func block() {

	var wg sync.WaitGroup

	clientA, err := c.NewBasicClient("localhost:10001",
		"/home/erik/dev/alice/tls.cert",
		"/home/erik/dev/alice/data/chain/bitcoin/simnet/",
		"simnet", c.MacFilename("admin.macaroon"))
	if err != nil {
		fmt.Println(err.Error())
	}
	clientB, err := c.NewLndServices(&c.LndServicesConfig{LndAddress: "localhost:10002",
		TLSPath:     "/home/erik/dev/bob/tls.cert",
		MacaroonDir: "/home/erik/dev/bob/data/chain/bitcoin/simnet",
		Network:     "simnet",

		Insecure: true})
	if err != nil {
		fmt.Println(err.Error())
	}

	ctx := context.Background()
	ctx2 := context.Background()
	ret, err := clientA.GetInfo(ctx, &lnrpc.GetInfoRequest{})
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(ret.BlockHash)

	//clientB.ChainNotifier
	invoiceStream, err := clientA.SubscribeInvoices(ctx2, &lnrpc.InvoiceSubscription{})
	if err != nil {
		fmt.Println(err.Error())
	}
	invoiceUpdates := make(chan *c.Invoice)
	streamErr := make(chan error, 1)

	wg.Add(1)

	go func() {
		defer wg.Done()
		defer close(invoiceUpdates)
		defer close(streamErr)

		for {

			rpcInvoice, err := invoiceStream.Recv()
			if err != nil {
				streamErr <- err
				fmt.Println("err1")
				return
			}
			invoice, err := unmarshalInvoice(rpcInvoice)
			if err != nil {
				streamErr <- err
				fmt.Println("err2")
				return
			}
			invoiceUpdates <- invoice
			fmt.Println("added invoice!!")
			fmt.Println(invoice.Amount)
		}

	}()

	res2, err := clientA.AddInvoice(ctx, &lnrpc.Invoice{Value: 10})
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(res2.PaymentRequest)

	inv := <-invoiceUpdates
	if inv.State != invpkg.ContractOpen {
		fmt.Println("no open invoice")
		return
	}

	// Get info
	/*info, err := clientB.Client.GetInfo(ctx)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(info)*/

	// Get channel
	channels, err := clientB.Client.ListChannels(ctx, true, true)
	if err != nil {
		fmt.Println(err.Error())
	}

	res3 := clientB.Client.PayInvoice(ctx, res2.PaymentRequest, 10, &channels[0].ChannelID)
	if err != nil {
		fmt.Println(err.Error())
	}
	outchan := <-res3
	fmt.Println(outchan)

	time.Sleep(1 * time.Second)

	inv = <-invoiceUpdates
	if inv.State != invpkg.ContractSettled {
		fmt.Println("no settled invoice")
		return
	}
	fmt.Println("Settled")

	wg.Wait()

}
