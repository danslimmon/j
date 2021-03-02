# j: a tool for thinking

`j` is a tool for streamlining thought. it aims, through the magic of the UNIX console, to minimize
the activation energy for recording thoughts.

as you work, thoughts pop into your head. maybe they're thoughts about the work at hand, like

> i better remember to remove that debug statement before committing this change
> (more?)

or maybe they're offshoots from what you're doing, that you want to follow up on later, like

> google whether there exists a tool that could do this task for me in the future
> (more?)

or maybe they're totally unrelated 

> this would be a cool scene for that short story i'm writing: ...

anyway, whatever you're thinking about, `j` helps you get it out of your head and back on task. back
to getting things done.

# how it works

`j` does a few things:

## Note-taking

`j ta` lets you put a thought down:

```
~/$ j ta     # `{t}hought {a}dd`
```

run that, and you're editing a new, empty markdown document. write down whatever you need to write
down, and write the file. when you exit your editor, `j` saves your thought and drops you back to
your shell.

You can see what thoughts you've saved:

```
~/$ j tl     # `{t}hought {l}ist
5m      
```

## Card management

`j cap` adds a pink card. `j cab` adds a blue card.

```
~/$ j cap measure water level in basement
~/$ j cab -d look into weird network traffic spikes
<editor opened for further description>
```

other things:
- review workflow: unsorted thoughts, constraint violations, untriaged cards
- thought to card
- tmux integration
- all the project management stuff that j currently does
- would be cool if you could a/b test stuff and get data (e.g. headphones or not, white noise or
    music, jazz or classical)
- mobile thingy?
- status view
- tag-activated plugins e.g. resolve asana

# order to build things in:

- `j ca{p,b,w}` add cards
- `j cs` show cards (each card has an ID and exists in a column)
- `j ce` edit card by ID (or, if no ID, the card currently in progress)
- a card is a markdown document
- when you're editing a card, you can add at the top `+done` to move it to done or `+blocked` to
    move it to blocked o cetera
- `j cd (+|-)N ID` move a card up or down n spaces in its column
- if you write `+doing`, it moves the card in the doing column and calls any hooks for that. you
    don't need to reopen the file - it doesn't move on the filesystem
- some way when editing a project file to add a "spend 60 minutes on this project" card to Today
- all the current functionality (`j pl` etc.)

# example workflow, to build:

```
export J_WORKSPACE="${HOME}/j_workspace"
j cap
<editor opens, user writes card>
j cab
<editor opens, user writes card>
j dp # {d}ay {p}lan
<editor opens to frontmatter-only file. user moves cards into today>
j w 1h
<1h timer starts>
<pink card gets moved to "in progress">
<editor opens pink card; user works pink card; user closes pink card>
<blue card gets moved to in progress>
<editor opens blue card>
```

while the editor is running, j is watching for writes. when it sees them, it commits and pushes

work order is:
- j exceptions
- white cards
- pink cards
- blue cards

# what's the simplest thing to do

- `j ta`:
  - cd into J_WORKDIR
  - open a blank file, from a template, in editor
  - when closed, move it into thoughts/to_review

then next is `j tr` to review thoughts

# architecture

so at the very bottom, we have `StorageClient`, which provides an API to read and write files that
are stored in the workspace.

```
type StorageClient interface {
  // Returns the Document at the given path, relative to J_WORKSPACE
  Get(path string) (Document, error)
  // Returns all Documents whose title starts with the given prefix
  GetByTitlePrefix(titlePrefix string, activeOnly bool) ([]Document, error)

  // Stores the given Document at the given path, overwriting anything already there.
  Put(path string, doc Document) error
}

type StorageConfig struct {
  // The base directory for storage. If Git is enabled, this must be a Git repo.
  BaseDir string
  Git GitConfig
}

/*
NewStorageClient returns a StorageClient whose base directory is baseDir.
*/
func NewStorageClient(c StorageConfig) {...}
```

# #9 abstract workspace

- make Workspace with tests
