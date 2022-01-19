#  Creating Protected Entity Types

The main object in Astrolabe is the _Protected Entity_.  The Go API for this is in
[../pkg/astrolabe/protected_entity.go](../pkg/astrolabe/protected_entity.go).  _Protected Entities_ are
managed by _Protected Entity Type Managers_ 
[../pkg/astrolabe/protected_entity_type_manager.go](../pkg/astrolabe/protected_entity_type_manager.g0).

To implement a new _Protected Entity_ type, define the type name, new types are recommended to be defined
with a reverse DNS style name for uniqueness, e.g. _com.aws.ebs_.  Implement a _Protected Entity_ and
_Protected Entity Type Manager_ for the type.

In order to test, you can create a CLI that will run the _Protected Entity_.  Create a constructor for
your _Protected Entity Type Manager_ that conforms to the _InitFunc_ function definition (`type InitFunc func(params map[string]interface{},
s3Config S3Config,
logger logrus.FieldLogger) (ProtectedEntityTypeManager, error)`)  Then create a CLI main.go in your
project that creates the CLI using the InitFunc you have defined in _addOnInits_

    package main

    import "github.com/vmware-tanzu/astrolabe/pkg/cmd"

    func main() {
        addOnInits := make(map[string]server.InitFunc)
		addOnInits["your_type"] = your_package.NewYourTypeProtectedEntityTypeManager

        cmd.CliCore(nil)
    }

This will allow you to exercise the _Protected Entity's_ functionality using the CLI commands.  You will need
a config dir as defined in [config_dir.md](config_dir.md) to initialize your Protected Entity Type Manager.
