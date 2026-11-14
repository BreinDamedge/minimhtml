Writing an mhtml parser for miniwoofer to use.

Currently its only functionality is to convert mhtml to html.

related content like images and stylesheets have their links overwritten and replaced with URIs.

What else would you like it to do?
- [ ] extract just the html portion
- [ ] extract outlinks
  - [ ] extract resource links (location of other segments)
  - [ ] extract links in html section (can use this later for overwriting outlinks to other stored documents you have)
- [ ] extract resources
- [ ] parse from mhtml -> html (for rendering) in under 2 seconds for 1000 samples
- [ ] test it on 10 documents
  - visually inspect outputs to make sure it worked
- [ ] make this a module that can be it's own thing and used for miniwoofer.

