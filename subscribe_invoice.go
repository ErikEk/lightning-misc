package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	c "github.com/lightninglabs/lndclient"
	invpkg "github.com/lightningnetwork/lnd/invoices"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lntypes"
	"github.com/lightningnetwork/lnd/lnwire"
)

// unmarshalInvoice creates an invoice from the rpc response provided.
func unmarshalInvoice(resp *lnrpc.Invoice) (*c.Invoice, error) {
	hash, err := lntypes.MakeHash(resp.RHash)
	if err != nil {
		return nil, err
	}

	invoice := &c.Invoice{
		Preimage:       nil,
		Hash:           hash,
		Memo:           resp.Memo,
		PaymentRequest: resp.PaymentRequest,
		Amount:         lnwire.MilliSatoshi(resp.ValueMsat),
		AmountPaid:     lnwire.MilliSatoshi(resp.AmtPaidMsat),
		CreationDate:   time.Unix(resp.CreationDate, 0),
		IsKeysend:      resp.IsKeysend,
		Htlcs:          make([]c.InvoiceHtlc, len(resp.Htlcs)),
		AddIndex:       resp.AddIndex,
		SettleIndex:    resp.SettleIndex,
	}

	for i, htlc := range resp.Htlcs {
		invoiceHtlc := c.InvoiceHtlc{
			ChannelID:     lnwire.NewShortChanIDFromInt(htlc.ChanId),
			Amount:        lnwire.MilliSatoshi(htlc.AmtMsat),
			CustomRecords: htlc.CustomRecords,
			State:         htlc.State,
		}

		if htlc.AcceptTime != 0 {
			invoiceHtlc.AcceptTime = time.Unix(htlc.AcceptTime, 0)
		}

		if htlc.ResolveTime != 0 {
			invoiceHtlc.ResolveTime = time.Unix(htlc.ResolveTime, 0)
		}

		invoice.Htlcs[i] = invoiceHtlc
	}

	switch resp.State {
	case lnrpc.Invoice_OPEN:
		invoice.State = invpkg.ContractOpen

	case lnrpc.Invoice_ACCEPTED:
		invoice.State = invpkg.ContractAccepted

	// If the invoice is settled, it also has a non-nil preimage, which we
	// can set on our invoice.
	case lnrpc.Invoice_SETTLED:
		invoice.State = invpkg.ContractSettled
		preimage, err := lntypes.MakePreimage(resp.RPreimage)
		if err != nil {
			return nil, err
		}
		invoice.Preimage = &preimage

	case lnrpc.Invoice_CANCELED:
		invoice.State = invpkg.ContractCanceled

	default:
		return nil, fmt.Errorf("unknown invoice state: %v",
			resp.State)
	}

	// Only set settle date if it is non-zero, because 0 unix time is
	// not the same as a zero time struct.
	if resp.SettleDate != 0 {
		invoice.SettleDate = time.Unix(resp.SettleDate, 0)
	}

	return invoice, nil
}

func SubscribeInvoice() {

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
		Insecure:    true})
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
