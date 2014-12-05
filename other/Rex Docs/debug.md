PACKAGE DOCUMENTATION

package debug
    import "dctech/rex/commands/debug"



FUNCTIONS

func Command_Registers(script *rex.Script, params []*rex.Value)
    Print the value of all internal registers and flags.

	debug:registers

    Returns unchanged.

func Command_Shell(script *rex.Script, params []*rex.Value)
    Break into the debugging shell. This command will provide an interactive
    shell until a DOS EOF char is found, on windows this may be simulated by
    pressing CTRL+Z followed by <ENTER>.

	debug:shell

    Returns the return value of the last command to be run.

func Command_Value(script *rex.Script, params []*rex.Value)
    Print information about a script value.

	debug:value value

    Returns unchanged.

