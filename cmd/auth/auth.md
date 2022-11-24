# MES Authorization System

## What you need in this system

1. Execution file (ask [KD M2100 MES Leader](mailto:m2100@kenda.com.tw) for the execution file)
2. Configuration file (in [YAML](https://zh.wikipedia.org/wiki/YAML))

### How to use this system

1. Write Configuration File.

    You will need a configuration file to specify where your data needed to be inserted and the list of employees which you want to give authorization.

    **Configuration file format**:

    ```yaml
    postgres:
        name:
        address:
        port:
        username:
        password:
        schema:
    action:
    ids:
        - EMPLOYEE_ID1
    ```

    **Configuration description**:

    | Object Name | Sub-Object Name | Data Type | Description |
    | --- | --- | --- | --- |
    | postgres | | struct | postgreSQL database settings |
    | | name | string | database name |
    | | address | string | database IP address |
    | | port | integer | database port |
    | | username | string | database username |
    | | password | string | database password |
    | | [schema](https://www.postgresqltutorial.com/postgresql-administration/postgresql-schema/#:~:text=What%20is%20a%20PostgreSQL%20schema,schema_name.object_name) | string | database specified schema |
    | action |  | string | Optional parameter where<br />"create": create new accounts<br />"delete": delete accounts<br />"reset": reset password<br />default: create new accounts |
    | ids | | []string | employee-ids which you want to operate |

2. Execution

    1. Open Terminal or Command Line Prompt

    2. Go to the path where there is the execution file

    3. Execute the program with configuration file as command line flag

     ```bash
     ./auth.exe  --config=${configuration-file-path}
     ```

### Example

if I have `auth.exe` and configuration file `config.yaml` in `C:\auth_example`

![folder](https://i.imgur.com/VFT9U3B.png)

![tree](https://i.imgur.com/MA9hJOR.png)

#### CMD

Open CMD

```cmd
Microsoft Windows [版本 10.0.19042.685]
(c) 2020 Microsoft Corporation。著作權所有，並保留一切權利。

C:\Users\tester>
```

change directory by `cd`

```cmd
C:\Users\tester>cd C:\auth_example

C:\auth_example>
```

run `auth.exe`

```cmd
C:\auth_example>.\auth.exe --config=config.yaml
Operation completed!
                                                    __----~~~~~~~~~~~------___
                                   .  .   ~~//====......          __--~ ~~
                   -.            \_|//     |||\\  ~~~~~~::::... /~
                ___-==_       _-~o~  \/    |||  \\            _/~~-
        __---~~~.==~||\=_    -_--~/_-~|-   |\\   \\        _/~
    _-~~     .=~    |  \\-_    '-~7  /-   /  ||    \      /
  .~       .~       |   \\ -_    /  /-   /   ||      \   /
 /  ____  /         |     \\ ~-_/  /|- _/   .||       \ /
 |~~    ~~|--~~~~--_ \     ~==-/   | \~--===~~        .\
          '         ~-|      /|    |-~\~~       __--~~
                      |-~~-_/ |    |   ~\_   _-~            /\
                           /  \     \__   \/~                \__
                       _--~ _/ | .-~~____--~-/                  ~~==.
                      ((->/~   '.|||' -_|    ~~-/ ,              . _||
                                 -_     ~\      ~~---l__i__i__i--~~_/
                                 _-~-__   ~)  \--______________--~~
                               //.-~~~-~_--~- |-------~~~~~~~~
                                      //.-~~~--\)

```

#### Powershell

change direction by `Set-Location`

```powershell
PS C:\Users\tester> Set-Location C:\auth_example
```

run `auth.exe`

```powershell
PS C:\auth_example> .\auth.exe --config=config.yaml
Operation completed!
                                                    __----~~~~~~~~~~~------___
                                   .  .   ~~//====......          __--~ ~~
                   -.            \_|//     |||\\  ~~~~~~::::... /~
                ___-==_       _-~o~  \/    |||  \\            _/~~-
        __---~~~.==~||\=_    -_--~/_-~|-   |\\   \\        _/~
    _-~~     .=~    |  \\-_    '-~7  /-   /  ||    \      /
  .~       .~       |   \\ -_    /  /-   /   ||      \   /
 /  ____  /         |     \\ ~-_/  /|- _/   .||       \ /
 |~~    ~~|--~~~~--_ \     ~==-/   | \~--===~~        .\
          '         ~-|      /|    |-~\~~       __--~~
                      |-~~-_/ |    |   ~\_   _-~            /\
                           /  \     \__   \/~                \__
                       _--~ _/ | .-~~____--~-/                  ~~==.
                      ((->/~   '.|||' -_|    ~~-/ ,              . _||
                                 -_     ~\      ~~---l__i__i__i--~~_/
                                 _-~-__   ~)  \--______________--~~
                               //.-~~~-~_--~- |-------~~~~~~~~
                                      //.-~~~--\)
PS C:\auth_example>
```

#### Bash

change directory by `cd`

```bash
cd c:/auth_example
```

run `auth.exe`

```bash
tester@computer MINGW64 /c/auth_example
$ ./auth.exe --config=config.yaml
Operation completed!
                                                    __----~~~~~~~~~~~------___
                                   .  .   ~~//====......          __--~ ~~
                   -.            \_|//     |||\\  ~~~~~~::::... /~
                ___-==_       _-~o~  \/    |||  \\            _/~~-
        __---~~~.==~||\=_    -_--~/_-~|-   |\\   \\        _/~
    _-~~     .=~    |  \\-_    '-~7  /-   /  ||    \      /
  .~       .~       |   \\ -_    /  /-   /   ||      \   /
 /  ____  /         |     \\ ~-_/  /|- _/   .||       \ /
 |~~    ~~|--~~~~--_ \     ~==-/   | \~--===~~        .\
          '         ~-|      /|    |-~\~~       __--~~
                      |-~~-_/ |    |   ~\_   _-~            /\
                           /  \     \__   \/~                \__
                       _--~ _/ | .-~~____--~-/                  ~~==.
                      ((->/~   '.|||' -_|    ~~-/ ,              . _||
                                 -_     ~\      ~~---l__i__i__i--~~_/
                                 _-~-__   ~)  \--______________--~~
                               //.-~~~-~_--~- |-------~~~~~~~~
                                      //.-~~~--\)

rockefel@DESKTOP-4RCB60P MINGW64 /c/auth_example
$

```
