package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
)

type KeyServerAddrType string

const keyServerAddr KeyServerAddrType = "serverAddr"

// w is the responseWriter directly while,
// r is a pointer to the http.Request

/*
For now, in both of our HTTP handlers, you use fmt.Printf to print when a request comes in for the handler function, then you use the http.ResponseWriter to send some text to the response body.

The http.ResponseWriter is an io.Writer, which means you can use anything capable of writing to that interface to write to the response body.

In this case, you’re using the io.WriteString function to write your response to the body.
*/

func getRoot(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	fmt.Printf("got / request\n", ctx.Value(keyServerAddr)) // this logs on to the server
	fmt.Printf("this is the request object: %v\n", r)       // this logs on to the server
	io.WriteString(w, "this is my website")                 // this writes the string to w
}

func getHello(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	fmt.Printf("got /hello request\n", ctx.Value(keyServerAddr)) //same here as well
	fmt.Fprintf(w, "Hello, you've requested: %s\n", r.URL.Path)  //same here as well
	fmt.Printf("this is the response Writer: %d\n", w)           //same here as well
}

// What is the Default Server Multiplexer?
// In Go, http.HandleFunc(...) registers the handler on the default multiplexer (http.DefaultServeMux), which is a built-in router that maps paths to handler functions.

// So internally, Go is doing something like:
// --> http.DefaultServeMux.HandleFunc("/", getRoot)

/*
func main() {
	http.HandleFunc("/", getRoot)
	http.HandleFunc("/hello", getHello)

	fmt.Println("server is listening on port : 80")

	err := http.ListenAndServe(":3333", nil) // The nil means: use the default ServeMux, which already knows the paths and handlers you registered.

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
*/

// some important points to remember: -->

/*
	In this section we just created a default server and routed the requests using the default server multiplexer. This is fine for local development. Lekin, PROD mei gand Fat jaayegi.

	Since the default mux is shared globally. If another package calls http.HandleFunc(...), it affects our app’s router too. That can lead to surprises.

	Now we will create our own
*/

/*
****************************************************************
****************************************************************
**********MULTIPLEXING********REQUEST********HANDLERS***********
****************************************************************
****************************************************************
 */

/*
	func main() {
		mux := http.NewServeMux()

		mux.HandleFunc("/", getRoot)
		mux.HandleFunc("/hello", getHello)

		err := http.ListenAndServe(":3000", mux)

		if errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("server closed\n")
		} else if err != nil {
			fmt.Printf("error starting server: %s\n", err)
			os.Exit(1)
		}

}
*/
func main() {
	/*
	****************************************************************
	****************************************************************
	******RUNNING******MULTIPLE******SERVERS******AT******ONCE******
	****************************************************************
	****************************************************************
	 */

	//  Sometimes you may want to customize how the server runs, or you may want to run multiple HTTP servers in the same program at once. For example, you may have a public website and a private admin website you want to run from the same program. Since you can only have one default HTTP server, you wouldn’t be able to do this with the default one.

	mux := http.NewServeMux()

	mux.HandleFunc("/", getRoot)
	mux.HandleFunc("/hello", getHello)

	ctx, cancelCtx := context.WithCancel(context.Background())

	serverOne := &http.Server{
		Addr:    ":3333",
		Handler: mux,
		BaseContext: func(l net.Listener) context.Context {
			ctx = context.WithValue(ctx, keyServerAddr, l.Addr().String())
			return ctx
		},
	}

	serverTwo := &http.Server{
		Addr:    ":4444",
		Handler: mux,
		BaseContext: func(l net.Listener) context.Context {
			ctx = context.WithValue(ctx, keyServerAddr, l.Addr().String())
			return ctx
		},
	}

	go func() {
		err := serverOne.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("server one closed\n")
		} else if err != nil {
			fmt.Printf("error listening for server one: %s\n", err)
		}
		cancelCtx()
	}()

	go func() {
		err := serverTwo.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("server two closed\n")
		} else if err != nil {
			fmt.Printf("error listening for server two: %s\n", err)
		}
		cancelCtx()
	}()

	<-ctx.Done()
}
