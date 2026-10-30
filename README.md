# Dodeyo

**Dokodemoyomou** is the easiest way to export your favorite novels as epub, pdf or html, mail them to your kindle or export them to any ereader and keep practicing your japanese wherever you are.  
Supercharge your japanese learning journey with free novels from the biggest japanese amateur novelist site **[syosetsu.com](https://syosetu.com/)**.  
Scan novels to find the jlpt level of the vocabulary used, is the vocabulary too hard? Add furigana for an easier read. Want a more japanese style, convert novels into a vertical format.  
The app keeps track of the novels you download if you only download 10 of 100 chapters the next time you download the app will download from where you left off.  

## Quick Start

### With go installed

```bash
go install github.com/jpabloVega/dodeyo
```
 
## Usage 

### Download a novel

You can choose the output of the file by using one of these flags:
- --epub
- --html
- --pdf

File created as epub:

``
dodeyo get [syosetsu novel url] --epub
``

After the first download you will be asked to introduce a folder to store the file, this will become your default folder, you can change it later using the config command.

#### Flags

You can add flags after a command to change its behaviour.  

You can use this flag on any command to get more information.

``
--help
``

Short version.  

``
-h
``

### Get command flags

``
dodeyo get [syosetsu novel url] [flags]
``

Flags come after the url, you can use multiple of them, however **--epub**, **--html** and **--pdf** are exclusive to each other.

##### Name
Define a name for the document, it defaults to the novels title.  
``
--name [document name]
``  
Short version.  
``
--n [document name]
``  

##### Start From Chapter
You can start from any chapter.  
``
--start [number]
``  
If this flag is omitted, it will choose using this hierarchy:
1. The last chapter you downloaded from this novel.
2. The index in the link you introduce.
3. The first chapter

This command is for serialized novels only.

##### Amount
Amount of chapters to download. Defaults to 10.  
They will be included in the same document.  
``
--amount [number]
``  
Short version.  
``
-a [number]
``

##### Furigana
Add kana reading to kanji.  
``
--furigana
``  

##### Vertical
Create novel usign vertical reading.  
``
--vertical
``  
Short version.
``
-v
``

##### Save to
Define a directory to save the output for the current execution, this doesnt change the default directory to define a default folder use the config command.  
``
--saveto [path]
``  
Short version.  
``
-s [path]
``

##### Mail to kindle
Send the results of the execution to your kindle via **Gmail**.  
If no mail is set this will ask for your mail information.  
``
--send
``  
You will need an **App Password** for your Gmail account, this password will be saved in the password administrator on your device.
You also need to add your Gmail account into your authorized emails in your amazon kindle configuration.

##### Pdf format
You can change the size of the title and body.  
Change title size, default is 16.  
``
--titlesize [number]
``  
Change body size, default is 14.  
``
--bodysize [number]
``  

### Convert command
Convert txt files into a ereader format. This command uses the same flags as the get command.  
With the exceptions of **amount** and **start**.  
``
dodeyo convert [filepath .txt] [flags]
``  
The txt file will be converted using these rules.

- The first line of a paragraph will become the title.
- Chapters are divided by empty line jumps.
- If the first line in the last paragraph is named **Credits** the next two lines will be the author and the link.

### Scan command
Scan a novel to know the jlpt level of the vocabulary.  
``
dodeyo scan [syosetsu novel url] [flags]
``  
This command shares the next flags with the get command.

- --amount -a
- --start

##### Verbose
By default scan prints the number of words on each jlpt level found in the novel.
You can use this command to see the full list of words it found.  
``
--verbose
``  
Short version.  
``
-v
``  

### History command
You can see your novel download history.  
``
dodeyo history
`` 
This prints the name, link, download date, and the next chapter to download.  
You can use this to get the link of a novel you liked.

### Config
``
dodeyo config [flags]
``  

##### Default path
Set the default path.
``
--defaultPath [path]
``  

##### Gmail and App password
Access the new email menu to change your mail or app password.
``
--email
``  

##### Kindle mail
Set your kindle email.  
``
--kindleEmail [kindle email]
``  

##### Proxy
Set a proxy.
``
--proxy [proxy address]
``  
If a proxy is set, the program will continue to use it.  
To stop this you can remove it.  
``
--removeProxy
``

##### Show
Prints the current configuration values.  
``
--show
``

### Update command
Downloads and installs the newest version of dodeyo.  
``
dodeyo update
``

## Motivation
I got a kindle as a tool to learn japanese, but after perusing the store I was a dissappointed to find a lack of japanese novels.
Most of them where locked in the japanese kindle store, and accessing required using vpns and other shaddy methods.
Online you can find an extensive amount of free novels but reading on a screen is not very nice for the eyes.
So I created dodeyo as a tool for myself, but if you find it useful feel free to use it.

## Contributing
This is an open source project so feel free to clone it.
