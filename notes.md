 a parser that can take mhtml documents and convert them to html.

The only translation I need is mht -> html

It would be nice to have it transformed to an AST or smth whatever it's called for the html library in go.

then you could modify links and all that down the line, but also, for serving the data I think we just send the text directly.

let's get an example mhtml document.
why? what do you want to do with it? I want to look at it. also pull up the mhtml rfc.

resources to gather:
- [ ] mhtml rfc
- [ ] example mhtml document

goal:
- [ ] parse an example mhtml document into html
- [ ] do this efficiently (Linear time perhaps?)
- [ ] do it in such a way that it is streamed so that large files aren't a problem

k take a look at an example document and try to write out the parsing steps in plain english.

# Observations
- multipart boundaries are followed by some attributes but they always end with `\n\n`
- 

# Potential Approaches
1. Find the multipart boundary string
1. Split the document on that boundary
  - this would give you sections of the document which you could then play with. It would probably be more convenient than whatever I'd have to do to do this in one single forward pass, but it wouldn't be as fast.
is one forward pass reasonable? probably not but why? if you're just trying to do text, then one forward pass is very reasonable I think but if you're trying to replace links there would probably have to be at least two. What links would you be replacing? excuse me I mean If I want to take the image blobs and put them back into their appropriate spots. 

How are the image blobs defined?
  - they are base64 encoded blobs. they have some headers, one of which tells you what
What is the identifier for where they're supposed to go?
  - content_location field under the multipart boundary when they're assigned as images.

> Executive Decision
> Build it so that it is nice and usable and also *works* before you try and do it all fast in one go.

so the way the document is set up it's got a bunch of blobs one after another and they each start/end with a multipart boundary. in this we have an html blob, a css blob, and images. maybe other types.
each blob section starts with a multipart boundary and some attributes/headers underneath that. these headers include: 

```
Some Examples:

------MultipartBoundary--TSwKjzLsKyDhNwmGTGbf8Eir6nDxtKtR2F5EdBosko----
Content-Type: image/png
Content-Transfer-Encoding: base64
Content-Location: https://gameprogrammingpatterns.com/images/observer-weasel-wielder.png

------MultipartBoundary--TSwKjzLsKyDhNwmGTGbf8Eir6nDxtKtR2F5EdBosko----
Content-Type: text/css
Content-Transfer-Encoding: quoted-printable
Content-Location: https://fonts.googleapis.com/css?family=Merriweather:400,400italic,700,700italic|Source+Code+Pro|Source+Sans+Pro:200,300,400,600,400italic,600italic|Rock+Salt

------MultipartBoundary--TSwKjzLsKyDhNwmGTGbf8Eir6nDxtKtR2F5EdBosko----
Content-Type: text/html
Content-ID: <frame-3D3108BEB8F8FFCB9854E8BE46066E6C@mhtml.blink>
Content-Transfer-Encoding: quoted-printable
Content-Location: https://gameprogrammingpatterns.com/observer.html
```

the first header does always seem to be the content-type. that's good.
then transfer-encoding then location n stuff if it has one.

got caught in the rabbit hole again. so what's all this useful for? how do I put it together. put it together.

# Steps
1. Get the multipart boundary string
1. Split the original text into it's multipart chunks
1. Convert those chunks of text into some struct
1. Go back through the html block for all the links that correspond to the chunks that we have from the document (since they aren't links anymore, we already have the data)
1. replace all the links with the data from the chunks
1. at this point you should have an html blob ready to go.

Does this seem like a good plan?
somewhat yes. the converting to a struct might not make the most sense, and the "go back" also raises an eyebrow.

so that is what I'd like to acomplish but how about this:
1. stream the input file
1. map the file: store the starting position of relavent data that we'll want to pull from the file later
  - read through and find the position of all multipart boundaries in the file
  - find links
1. now, using the map start building the output string.
  1. plop the html in up to the head probably
  1. plop in the css blob (might need to do link things?)
  1. keep plopping html content until you hit a link that needs overwriting
    - overwrite it
  1. repeat until you're out of the html

I'm sorta liking that but I don't feel comfortable that it makes 100% of the sense that it needs to for me to implement it.

walk me through it again in words/steps like that and then do like a mock run of it on example doc.

# conceptual walkthrough
## Indexing Step
okay so you can picture this as a reader head reading over the input string
it's going to keep stepping forward and when it finds a string that it recognizes, it'll plop down a mark on a list or something of that string it was looking for and a position where it started.
that's the sort of mapper interface.
you say "look for these patterns" and it'll tell you their positions in the file. that sounds a lot like regex. you know, it probably is regex. can regex give you the position of a match? Can it look for multiple things in one sweep? It can, but rei says it's usually better to make your own matcher and I'm inclined to believe them.

Okay so we make a `MatchIndexer` or something, whatever we choose to call it, and then it gets a "these are the patterns you're looking for" and when you find one, put a tick of where it was.
It could take a list of patterns and then use those as keys to a map, and the map gives you a list of starting indexes of that pattern. might also be useful to have the data structure be sorted by the position of the found objects but lets leave that for later ig.

sorry I got distracted singing for a little bit. where was I? MatchIndexer builds an index of the file (specifically of boundaries & links). Then for each boundary, you should dispatch some reader that can tell you what type of stuff is there using the headers
  > or, when the indexer finds a match for a mulitpart boundary, just immediately figure out what type of content it is. You could seek the next `:` and from there seek the previous `\n` and between those two symbols should be your attribute name. that'd probably work as a paradigm.

> side note for when you talk about replacing the content of links: all the `link rel=3D"www.content.foo"`, specifically the `=3D` part is because the encoding is `quoted-printable` so keep an eye out for if that will cause you problems later. might also just be worth reading a bit about the quoted-printable format.

## Building The Output
...

I'm going to pause here and actually do a bit of coding since I'm in go land and I want to make sure that my understanding of what I want to do aligns with what I can do. Switching keyboards now :\[

What coding would I like to do?
- [x] build a prototype of the MatchIndexer which only matches for the multipart boundary
- [x] build a function that can recognize and extract a multipart boundary

both of these are done.
I've been using the bufio scanner. It works fine except that I can't move backward. everything I need to parse should be on one line? nope! but kind of yes. Soft line breaks (`=\n`) are a thing because of the quoted-printable format again (line length max of 76 chars I think)

next up things will probably be to recognize what each section is/contains, and recognize & track links.
- [ ] recognize sections in code
  - when you find a multipart section spin off a thread (or a go routine) that can go parse this section (which for now just means read the header)

well okay. The tape analogy isn't quite right.

also let's answer the question of how you plan to replace the values.
you could replace the bare text which might work, you could also build pass it to go's html library to build a dom, then search the dom for links and replace them with blob data.
- [x] look into decoding the quoted-printable text
  - https://pkg.go.dev/mime/quotedprintable

use this to make a quoted-printable decoder and decode the html blob from the multipart-related doc.

what would you like this parser to be able to do?
- [ ] Parse an mhtml document into html
- [ ] replace links in the html body that are present in the mhtml document with the data that those links encode for if present

### Finding Links
...

good morning, it's a new day. today, I'm going to actually implement something.

we're going to do it poorly today and test how slow it is.

1. find the separator
1. loop through and for each separator
  1. make a struct
  1. read the headers
  1. plop the blob into memory
1. write the html to a new file (no image or css replacement yet)

