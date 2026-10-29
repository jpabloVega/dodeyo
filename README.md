# Dodeyo

Supercharge your japanese learning journey with free novels from the biggest japanese amateur novelist site **[syosetsu.com](https://syosetu.com/)**.
Export your favorite novels as epub, pdf or html. Email them to your kindle or export them to any ereader and keep practicing your japanese wherever you are.
Scan novels to find the jlpt level of the vocabulary used, is the vocabulary too hard? Add furigana for an easier read. Want a more japanese style, convert novels into a vertical format.
The app keeps track of the novels you download if you only download 10 of 100 chapters the next time you download the app will download from where you left off.

## Quick Start

### With go installed

``bash
go install github.com/jpabloVega/dodeyo
``
 
## Usage 

### Download a novel

You can choose the output of the file by using one of these flags:
- --epub
- --html
- --pdf

``bash
dodeyo get [syosetsu novel url] --epub [flags]
``

After the first download you will be asked to introduce a folder to store the file, this will become your default folder, you can change it later using the config command.

#### Get command flags

Flags come after the url, you can use multiple of them, however **--epub**, **--html** and **--pdf** are exclusive to each other.
You can use the next flag to get addicional information of any command:

``bash
--help
short version
-h
``
##### Name

Define a name for the document, it defaults to the novels title.

``bash
--name [document name]
short version
--n [document name]
``


##### Start From

When you download a novel it will start from the chapter of the link you give or the first chapter if you introduce the link from the index page of the novel.
If you downloaded this novel previously it will start from the last chapter downloaded.
You can bypass this by using this command.

``bash
--start [number]
``

Order of operation from most to least priority.

1. --start flag
2. last downloaded chapter
3. index of the given link

For one chapter novel this does nothing.

##### Amount

The number of chapters to download from a novel. Defaults to 10.

``bash
--amount [number]
short version
-a [number]
``

##### Furigana

Add kana reading to kanji.

``bash
--furigana
``

##### Vertical

Create novel usign vertical reading.

``bash
--vertical
short version
-v
``

##### Save to

Define a directory to save the output for the current execution, to define a default folder use the config command.

``bash
--saveto
short version
-s
``

##### Mail to kindle

Send the results of the execution to your kindle via **Gmail**.
If no mail is set this will ask for your mail information.

``bash
--send
``

You will need an **App Password** for your Gmail account, this password is saved in the password administrator of your device.
You will also need to add your Gmail account into your authorized emails in your amazon kindle configuration.

##### Pdf format

You can change the size of the title and body.

Change title size, default is 16.

``bash
--titlesize [number]
``

Change body size, default is 14.

``bash
--bodysize [number]
``

#### Convert command

Convert txt files into a ereader format. This command uses the same flags as the get command.
With the exceptions of **amount** and **start**.

``bash
dodeyo convert [filepath .txt] [flags]
``

The txt file will be converted using these rules.

- The first line of a paragraph will become the title.
- Paragraphs are divide by empty line jumps.
- If the last paragraph title is named Credits the next two lines will be the author and the link.

#### Scan command

Scan a novel from its link, to know the jlpt level of the vocabulary.

``bash
dodeyo [syosetsu novel url] [flags]
``

This command shares the next flags with the get command.

- --amount -a
- --start

##### Verbose

By default scan prints the number of words on each jlpt level found in the novel.
You can use this command to see the words it found.

``bash
--verbose
short version
-v
``

#### History command

You can see your novel download history.

``bash
dodeyo history
`` 

This prints the name, link, download date, and last download.
You can use this to continue downloading a novel you liked.

#### Config

``bash
dodeyo config [flags]
``

You can change your config by using these flags after the config command.

##### Default path

Set the default path.

``bash
--defaultPath [path]
``

##### Gmail and App password

Access the new email menu to change your gmail or app password.

``bash
--email
``

##### Kindle mail

Set your kindle mail.

``bash
--kindleEmail [kindle email]
``

##### Proxy

Set a proxy for novel scrapping.

``bash
--proxy
``

If a proxy is set, the program will continue to use it.
To stop this you can remove it.

``bash
--removeProxy
``

##### Show

Prints the current configuration values.

``bash
--show
``

#### Update command

Downloads and installs the newest version of dodeyo.

``bash
dodeyo update
``

## Motivation

I got a kindle as a tool to learn japanese, but after perussing the store I was a dissappointed to find a really small of japanese novels.
Most of them where locked in the japanese kindle store, and accessing required using vpns and other shaddy methods.
Online you can find an extensive amount of free novels but reading on a screen is not very nice for the eyes.
So I created dodeyo as a tool for myself, but if you find it useful feel free to use it. 

## Contributing

This is an open source project so feel free to clone and play with it.
