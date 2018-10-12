=======
goflick
=======

Package provides convenient interface for querying `Flick Electric`_ API.

Currently only the NeedlePrice is supported.

Installation
============

Fetch the package::

  go get github.com/rusq/goflick
  

Run tests::
  cd $GOPATH/src/github.com/rusq/goflick

go test

Usage
=====

.. code-block:: go

  package main

  import (
      "fmt"

      "github.com/rusq/goflick"
  )

  func main() {
      flick, err := goflick.NewConnect("email", "password")
      if err != nil {
          panic(err)
      }
      // get current price
      price, err := flick.GetPrice()
      if err != nil {
          panic(err)
      }
      fmt.Println(price)
  }
   

.. _`Flick Electric`: https://www.flickelectric.co.nz/
