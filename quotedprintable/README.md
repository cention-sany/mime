# quotedprintable

## Introduction

These files shamelessly copied from golang 1.5.1 standard library.
It is meant to be used in 1.4.2 development environment.

## Features

Unlike golang stdlib, this forked version relax the quotedprintable reader.
For example, if a bad email client stated the header quotedprintable content, 
but the email content has the '=' sign that is not escape with =3D, this 
forked version will not output error, instead will just ignore decode it, 
and print out according to whatever it receives.
