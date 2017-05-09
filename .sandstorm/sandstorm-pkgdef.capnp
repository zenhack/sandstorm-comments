@0xb933773aee188470;

using Spk = import "/sandstorm/package.capnp";
# This imports:
#   $SANDSTORM_HOME/latest/usr/include/sandstorm/package.capnp
# Check out that file to see the full, documented package definition format.

const pkgdef :Spk.PackageDefinition = (
  id = "14am1tdyfgxp6018m5kr10nfvf72yvgdajqsz5tkhts7g1uz3tuh",
  manifest = (
    appTitle = (defaultText = "Sandstorm Comments"),
    appVersion = 0,
    appMarketingVersion = (defaultText = "0.0.0"),
    actions = [
      ( nounPhrase = (defaultText = "comment manager"),
        command = .myCommand
      )
    ],
    continueCommand = .myCommand,
    metadata = (
      icons = (
        # Various icons to represent the app in various contexts.
        #appGrid = (svg = embed "path/to/appgrid-128x128.svg"),
        #grain = (svg = embed "path/to/grain-24x24.svg"),
        #market = (svg = embed "path/to/market-150x150.svg"),
        #marketBig = (svg = embed "path/to/market-big-300x300.svg"),
      ),

      website = "https://github.com/zenhack/sandstorm-comments",
      codeUrl = "https://github.com/zenhack/sandstorm-comments",
      license = (openSource = apache2),
      categories = [webPublishing],
      author = (
        contactEmail = "ian@zenhack.net",
        pgpSignature = embed "pgp-signature",
      ),

      pgpKeyring = embed "pgp-keyring",

      #description = (defaultText = embed "path/to/description.md"),
      # The app's description in Github-flavored Markdown format, to be displayed e.g.
      # in an app store. Note that the Markdown is not permitted to contain HTML nor image tags (but
      # you can include a list of screenshots separately).

      shortDescription = (defaultText = "Blog comment manager"),
      screenshots = [
        # Screenshots to use for marketing purposes.  Examples below.
        # Sizes are given in device-independent pixels, so if you took these
        # screenshots on a Retina-style high DPI screen, divide each dimension by two.

        #(width = 746, height = 795, jpeg = embed "path/to/screenshot-1.jpeg"),
        #(width = 640, height = 480, png = embed "path/to/screenshot-2.png"),
      ],
      #changeLog = (defaultText = embed "path/to/sandstorm-specific/changelog.md"),
      # Documents the history of changes in Github-flavored markdown format (with the same restrictions
      # as govern `description`). We recommend formatting this with an H1 heading for each version
      # followed by a bullet list of changes.
    ),
  ),

  sourceMap = (
    searchPath = [
      ( sourcePath = "." ),  # Search this directory first.
      ( sourcePath = ".."),  # Then the root of the repository.
      ( sourcePath = "/",    # Then search the system root directory.
        hidePaths = [ "home", "proc", "sys",
                      "etc/passwd", "etc/hosts", "etc/host.conf",
                      "etc/nsswitch.conf", "etc/resolv.conf" ]
      )
    ]
  ),

  fileList = "sandstorm-files.list",
  # `spk dev` will write a list of all the files your app uses to this file.
  # You should review it later, before shipping your app.

  alwaysInclude = [],
  # Fill this list with more names of files or directories that should be
  # included in your package, even if not listed in sandstorm-files.list.
  # Use this to force-include stuff that you know you need but which may
  # not have been detected as a dependency during `spk dev`. If you list
  # a directory here, its entire contents will be included recursively.

  bridgeConfig = (
    viewInfo = (
      permissions = [
        (
          name = "admin",
          title = (defaultText = "admin"),
          description = (
            defaultText = "Change settings and moderate comments"
          ),
        ),
        (
          name = "post-moderated",
          title = (defaultText = "Post moderated comments"),
          description = (
            defaultText = "Post moderated comments"
          ),
        ),
        (
          name = "post-unmoderated",
          title = (defaultText = "Post unmoderated comments"),
          description = (
            defaultText = "Post unmoderated comments"
          ),
        ),
      ],
      roles = [
        (
          title = (defaultText = "administrator"),
          permissions  = [true, true, true],
          verbPhrase = (defaultText = "change settings and moderate comments"),
          description = (
            defaultText = "Administrators may change settings and moderate comments."
          ),
        ),
        (
          title = (defaultText = "commentor"),
          permissions  = [false, true, false],
          verbPhrase = (defaultText = "post (moderated) comments"),
          description = (defaultText = "Commentors can post (moderated) comments"),
        ),
        (
          title = (defaultText = "viewer"),
          permissions = [false, false, false],
          verbPhrase = (defaultText = "view comments"),
          description = (defaultText = "Viewers can view comments."),
        ),
        (
          title = (defaultText = "trusted commentor"),
          permissions  = [false, true, true],
          verbPhrase = (defaultText = "post (unmoderated) comments"),
          description = (defaultText = "Trusted commentors can post (unmoderated) comments"),
        ),
      ],
    ),
  ),
);

const myCommand :Spk.Manifest.Command = (
  # Here we define the command used to start up your server.
  argv = ["/sandstorm-http-bridge", "8000", "--", "/app.stripped"],
  environ = [
    # Note that this defines the *entire* environment seen by your app.
    (key = "PATH", value = "/usr/local/bin:/usr/bin:/bin"),
    (key = "SANDSTORM", value = "1"),

    (key = "DB_PATH", value = "/var/db.sqlite3"),
    (key = "STATIC_ASSETS", value = "/static"),
    (key = "TEMPLATE_DIR", value = "/templates"),
    (key = "SCHEMA_FILE", value = "/schema.sql"),
  ]
);

# vim: set ts=2 sw=2 et :
