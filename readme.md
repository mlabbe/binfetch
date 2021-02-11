# Binfetch #

A minimalist, secure package manager for your intranet.

**Problem:** You have project exes that need to be securely distributed to company-internal end users and be non-interactively processed.  You cannot host publicly.

**Solution:** Upload your build products to an S3 store, and use `binfetch` to retrieve the latest executable for your os and architecture.


## Example Usage ##

``` {.text}
>binfetch ls      # list all projects


  Available projects:
       project_a
          - master
          - devel
        
       project_b
          - master
```


``` {.text}
>binfetch get project_a      # download the latest archive for your OS for master branch

  Found compatible archive built at 2021-02-09 19:00:09 -0800 PST
  Downloading project_a-linux-amd64-1.0.0-pre.tar.gz ...
  success.
```

`binfetch` detects the OS and architecture, and downloads the latest archive.

## How it works ##

An S3 store is to be configured with objects in the following structure:

`/$projname/$branch/$epoch__$any_tag_at_all/...`
   
 - `$projname` is a unique name for the project that differentiates it 
 - `$branch` is the name of a branch that produced the archive
 - `$epoch` is a unix timestamp indicating the last change to the build.
 - `$any_tag_at_all` is any helpful identifier, such as a git SHA1

The archive filename must specify the os and architecture. 

Valid OS strings:

 - win32
 - linux
 - macos
 
Valid architecture strings:
 
 - x86
 - x64
 - arm64
 
## Installation and Config ##

 1. `go get -u github.com/mlabbe/binfetch/cmd/binfetch`

 2. `aws configure` to specify an AWS profile with S3 Read Access.

 3. See `configs/` in this repo for a sample config.
 
 4. Create an S3 bucket and populate it (see "How it works")
 
 
