# Software Development Kit

The main goal of the Software Development Kit (SDK), is to create a sensible and intuitive abstraction to the [Application Programming Interface (API)](https://www.ibm.com/topics/api) of the Proxmox Virtual Environment (PVE).

When this document refers to the SDK, it references the code in the [/proxmox](../../../proxmox/) directory.

## Table of Contents

- [Software Development Kit](#software-development-kit)
  - [Table of Contents](#table-of-contents)
  - [Data Structures](#data-structures)
  - [Mapping between SDK and API](#mapping-between-sdk-and-api)
    - [Mapping to API](#mapping-to-api)
      - [Example Mapping to API](#example-mapping-to-api)
    - [Mapping to SDK](#mapping-to-sdk)
      - [Example Mapping to SDK](#example-mapping-to-sdk)
  - [Validator Pattern](#validator-pattern)
    - [Implementation of the Validator Pattern](#implementation-of-the-validator-pattern)
      - [Explicit Validator](#explicit-validator)
      - [Implicit Validator](#implicit-validator)
    - [Validator Errors](#validator-errors)
      - [Error Constant](#error-constant)
      - [Error Function](#error-function)
  - [Safe and Unsafe Functions](#safe-and-unsafe-functions)
    - [Safe function](#safe-function)
    - [Unsafe function](#unsafe-function)
  - [Public and Private](#public-and-private)
  - [Standardized Interfaces](#standardized-interfaces)
  - [Pure and Impure Functions](#pure-and-impure-functions)
    - [Pure Functions](#pure-functions)
    - [Impure Functions](#impure-functions)
  - [Testing](#testing)
  - [Versioning](#versioning)
  - [Anti Patterns](#anti-patterns)

## Data Structures

In the upcoming section, we will delve into the details of the two primary data structures at play:

1. SDK Data Structure: The software development kit (SDK) employs a [`struct`](https://go.dev/ref/spec#Struct_types) as it's core data structure.
2. API Data Format: Data obtained from the API arrives in [JavaScript Object Notation (JSON)](https://www.json.org/json-en.html) format and will be utilized in its raw form, represented as a [`map[string]interface{}`](https://bitfieldconsulting.com/golang/map-string-interface).

**Note:** When defining configurations for an object, it's essential to prefix only the top-level configuration with `Config`.
This practice is designed to enhance the user experience when utilizing code completion features in your [Integrated Development Environment (IDE)](https://www.codecademy.com/article/what-is-an-ide).

## Mapping between SDK and API

The main objective of the SDK is to convert the data structures of the SDK to and from the API calls that PVE understands.
These processes are referred to as `Mapping to API` and `Mapping to SDK`.
In the following chapters, both of these topics will be covered in greater detail.
In most cases the data structure used by the SDK and API are wildly different.
The reason for this is that the SDK is supposed to improve the developer experience.

### Mapping to API

The SDK needs to converts it's data structure to something the API understands.
In [Example Mapping to API](#example-mapping-to-api) a code example is given of how to convert the SDK data structure to the API input.
The Rules of the logic for mapping to the API are:

- The SDK config should have a private function named `mapToApi()` that is responsible for converting the JSON for the API from the SDK config.
- When Converting from the SDK to the API small and simple transformations may be done in the top level objects `mapToApi()` function. For more complex and bigger transformations dedicated functions should be used.

#### Example Mapping to API

```json
{
    "itemunit": "56", // this is always a number represented as a string
    "flag": "true,1,false", //this are always 3 values and are always in the same order
    "configdescription": "long comment" //this is always a string
}
```

```go
func main(){
    params := ConfigExample{
        ID:      56,
        Comment: "long comment",
        Flags: ExampleFlags{
            A: true,
            B: true,
            C: false,
        },
    }.mapToApi()
}

type ExampleID uint

// api key 'flag' was given a custom to type to improve clarity.
type ExampleFlags struct {
    A bool
    B bool
    C bool
}

func (flags ExampleFlags) mapToApi() string {
    // omitted for brevity.
}

// all root configs should be prefixed with 'Config'
type ConfigExample struct {
    ID      ExampleID
    Comment string
    Flags   ExampleFlags
}

func (config ConfigExample) mapToApi() (params map[string]interface{}) {
    // when transformations are simple they can be done in the main `mapToApi()` function.
    params["itemunit"], _ = strconv.Atoi(string(config.ID))
    params["configdescription"] = config.Comment
    // when transformations are complex they require their own `mapToApi()` function.
    params["flag"] = config.Flags.mapToApi()
    return
}
```

### Mapping to SDK

The SDK needs to convert the API response to something the SDK understands.
In [Example Mapping to SDK](#example-mapping-to-sdk) a code example is given of how to convert the API response to the SDK data structure.
The Rules of the logic for mapping to the SDK are:

- The SDK config should have a private function named `mapToSdk()` that is responsible for converting the JSON from the API to the SDK config.
- When Converting from the API to the SDK small and simple transformations may be done in the top level objects `mapToSdk()` function. For more complex and bigger transformations dedicated functions should be used.

#### Example Mapping to SDK

```json
{
    "itemunit": "56", // this is always a number represented as a string
    "flag": "true,1,false", //this are always 3 values and are always in the same order
    "configdescription": "long comment" //this is always a string
}
```

```go
func main(){
    params := map[string]interface{}{
        "itemunit": "56",
        "flag": "true,1,false",
        "configdescription": "long comment",
    }
    config := ConfigExample{}.mapToSDK(params)
}

type ExampleID uint

// api key 'flag' was given a custom to type to improve clarity.
type ExampleFlags struct {
    A bool
    B bool
    C bool
}

func (ExampleFlags) mapToSDK(rawFlags string) flags ExampleFlags {
    // omitted for brevity.
}

// all root configs should be prefixed with 'Config'
type ConfigExample struct {
    ID      ExampleID
    Comment string
    Flags   ExampleFlags
}

func (ConfigExample) mapToSDK(params map[string]interface{}) (config ConfigExample) {
    if itemValue, isSet := params["itemunit"]; isSet {
        tmpID,_:=strconv.Itoa(itemValue.(string))
        config.ID = ExampleID(tmpID)
    }
    if itemValue, isSet := params["configdescription"]; isSet {
        config.Comment = itemValue.(string)
    }
    if itemValue, isSet := params["flag"]; isSet {
        config.Flags = ExampleFlags{}.mapToSDK(itemValue.(string))
    }
    return
}
```

## Validator Pattern

The SDK makes heavy use of the [Validator pattern](https://eddieabbondanz.io/post/software-design/validator-pattern).
The reason for using this pattern instead of the [Getter Setter pattern](https://stackoverflow.com/questions/565095/are-getters-and-setters-poor-design-contradictory-advice-seen) is that it increases clarity and accessibility of the SDK.
This is achieved by utilizing objects with [public](https://yourbasic.org/golang/public-private/) properties, resulting in an increased number of `Validate()` functions for all the public properties.
In the [following](#implementation-of-the-validator-pattern) chapter the implementation of the validator pattern will be discussed.

### Implementation of the Validator Pattern

Implementing the Validator pattern for the SDK comes down to two principles:

- Custom types for a lot of values.
- A validate function for each [custom type](https://appdividend.com/2022/09/08/golang-custom-type-declaration/).

The reason for the custom types is that in a [statically typed](https://www.baeldung.com/cs/statically-vs-dynamically-typed-languages) language the variable type implies the constraint on the variables value.

The [implicit example](#implicit-validator) implements the Validator pattern without custom types, and the [explicit example](#explicit-validator) implements the Validator pattern with custom types.
The biggest difference between the [implicit example](#implicit-validator) and [explicit example](#explicit-validator), is that the [explicit example](#explicit-validator) explicitly connects the `PoolName.Validate()` function to the `PoolName` type.
In contrast to the [implicit example](#implicit-validator) which implicitly connects the `StringValidate()` function to the `string` type. For the above stated reason the [explicit example](#explicit-validator) is to be used.

Below are two examples:

#### Explicit Validator

```go
// Only alphanumerical characters, with a max lent of 20.
type PoolName string

func (valueToValidate PoolName)Validate() (err error) {
    return
}

func Example() {
var exampleName PoolName
err := exampleName.Validate()
}
```

#### Implicit Validator

```go
func StringValidate(valueToValidate string) (err error) {
    // Validation logic goes here.
    return
}

func Example() {
// Any charters and all lengths.
var exampleString string
exampleString = "This string is allowed to be anything like $TR!NG and even emoji 😮"
err := StringValidate(exampleString)
}
```

### Validator Errors

When dealing with validation errors, it is crucial for the SDK to respond with a clear and informative error message.
This error message should not only convey the direct error message received from the API but also offer a comprehensible explanation of the encountered issue.
Furthermore, it should allow for distinguishing between various error types.
To accomplish this, we make use of two key mechanisms:

- [Error constants](#error-constant)
- [Error functions](#error-function)

The purpose of employing [error constants](#error-constant) is to provide a straightforward and concise error message that can be conveniently referenced.
Conversely, [Error functions](#error-function) come into play when the error message varies depending on the specific item being validated.
To maintain simplicity, we strive to minimize the number of [Error functions](#error-function).

#### Error Constant

Error constants come into play when the error message remains the same for all items.
These constants should adhere to the following naming convention:

- Start with the type, followed by `_Error_`, and then the error name.

This naming convention facilitates error grouping by type, ultimately enhancing the developer experience, particularly when using code completion within an integrated development environment (IDE).
If we were to adopt the naming convention of Error_ followed by the error name, the IDE would consolidate all errors, making it more challenging to locate the specific error you seek.

```go
type UserName string

const UserName_Error_Invalid string = "username is invalid"
```

#### Error Function

Error functions are utilized when the error message contains variable information, such as the name of the item being validated. These functions should adhere to the following naming convention:

- They should be integrated into the corresponding class or type they are validating, and their names should start with `Error_`, followed by the error name.

```go
type UserName string

func (u UserName) Error_InvalidUsername() error {
    return fmt.Errorf("username '%s' is invalid", u)
}
```

## Safe and Unsafe Functions

Safe and unsafe functions serve as key components for interacting with the API. The rationale behind their existence lies in the trade-off between validating input for correctness and maintaining performance efficiency. For instance, consider the validation of whether a given name adheres to the API's naming conventions. While this validation is essential, performing it with every API interaction can significantly impact performance.

In certain scenarios, input validation may necessitate multiple API calls. Explicit unsafe functions are introduced to convey to developers that input validation is not carried out automatically, placing the responsibility for validation squarely on the developer. By offering both safe and unsafe functions, the SDK becomes more developer-friendly, as it allows developers to make informed choices between performance optimization and safety without the need to create their customized SDK version.

### Safe function

By default, all functions are considered safe functions, meaning that input validation occurs before the function execution. These functions are expected to provide developers with clear error messages when the input is invalid.

```go
type User struct

func (user *User) Create(password string) (err error) {
    err = user.Validate()
    if err != nil {
        return
    }
    // omitted for brevity.
    // more validation logic.
    err = user.Create_Unsafe(password)
    return
}
```

### Unsafe function

Unsafe functions, on the other hand, do not validate the input. These functions directly transmit the provided data to the API. They should be employed when the input is already validated, either by the developer or by a safe function.

```go
type User struct

// No validation is done here.
func (user *User) Create_Unsafe(password string) (err error) {
    // omitted for brevity.
    return
}
```

## Public and Private

In crafting our SDK, we prioritize developer-friendliness, aiming to make it as intuitive as possible.
To realize this goal, we employ the principle of [Public and Private functions](https://yourbasic.org/golang/public-private/).
This principle serves to distinguish between functions intended for developer use and those that are not.
The following rules govern the categorization of functions as public or private:

- Functions that provide a clear useful abstraction to the API should be made public.
- Functions that do some internal work and are not intended for direct developer interaction and should be designated as private.
- Functions that do not provide a clear useful abstraction to the API should be designated as private.

Public functions should follow this guide as much as possible, to realize a homogenous developer experience.

It is important to note that every public function becomes a critical point of reliance for some software.
As such, it is imperative to uphold backward compatibility when making changes. See [Versioning](#versioning) for more information.

## Standardized Interfaces

Standardized interfaces are key in software development for consistent functionality exposure. They enhance code understanding, usage, and system interoperability.

If the Go standard library has an interface for existing functionality, use it. It's designed with best practices, well-documented, and tested. Ignoring it can lead to complexity and compatibility issues.

## Pure and Impure Functions

In software development, functions are categorized into two main types: [pure functions](https://en.wikipedia.org/wiki/Pure_function#Pure_functions) and [impure functions](https://en.wikipedia.org/wiki/Pure_function#Impure_functions).
Understanding the distinction between these two types is crucial for building robust and maintainable software systems.

### Pure Functions

Pure functions are a fundamental concept in programming.
They have the following characteristics:

1. [Idempotent](https://en.wikipedia.org/wiki/Idempotence): Pure functions always produce the same output for the same input, and they have no hidden side effects.
2. No Side Effects: They don't modify external variables or perform I/O operations.
3. Referential Transparency: You can replace a function call with its result without changing the program's behavior.
4. Thread Safety: They are inherently thread-safe and can be used in concurrent or parallel programming.

Benefits:

- Predictable and easy to reason about.
- Simplify testing and debugging.
- Promote code reusability and modularization.

Example:

```go
func add(a, b int) int {
    return a + b
}
```

### Impure Functions

Impure functions have one or more of the following characteristics:

1. Side Effects: They may modify external variables, change global state, or perform I/O operations.
2. Non-Deterministic: Impure functions can produce different results for the same input based on external factors.
3. Dependent on External State: They rely on external state or shared resources.

Impure functions are not inherently "bad" but should be managed carefully to avoid unexpected behavior or bugs in your code.

Example:

```go
func writeToDisk(filename string, data []byte) error {
    file, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer file.Close()
    _, err = file.Write(data)
    return err
}
```

## Testing

Testing plays a crucial role in the development process.
It is essential that all public [pure](#pure-functions) functions have automated tests.
Because of this not every [pure](#pure-functions) private function needs to be tested as their functionality is implicitly tested by the public [pure](#pure-functions) functions.
In cases of complex [pure](#pure-functions) private functions, it is recommended to write tests for them.

**Note:** For dealing with the public [impure](#impure-functions) functions we use the integration tests from the [CLI tool](cli.md).

Below you'll find an example of a test for a public [pure](#pure-functions) function.
This test is for the `UserID.Validate()` function.
Every test should incorporate an inline-defined variable named `tests`, which should be an array of structs. The rationale behind this struct array is it's flexibility in expanding the test cases.
Each struct must encompass, at a minimum, the following attributes:

- `name`: The name of the test.
- `input`: The input for the function.
- `output`: The expected output of the function.

```go
func Test_UserID_Validate(t *testing.T) {
    tests:= []struct {
        name string
        input UserID
        output error
    }{
        {name: "Valid ID",
            input: UserID(1),
            output: nil,
        },
        {name: "Invalid ID",
            input: UserID(0),
            output: errors.New(UserID_Error_Invalid),
        },
    }
    for _, test := range tests {
        t.Run(test.name, func(t *testing.T) {
            require.Equal(t, test.output, test.input.Validate())
        })
    }
}
```

## Versioning

**Note: Until an official release is made, the SDK is considered to be in a state of constant development. As such, backward compatibility is not guaranteed.**

For versioning, we employ the [Semantic Versioning](https://semver.org) standard. This standard is designed to convey the impact of a new version on existing implementations.

## Anti Patterns

In this chapter, we will discuss some anti-patterns that should be avoided when developing the SDK.
Theses anti-patterns are mostly related to patterns in older code within this project and should be avoided in new code.

- **Do not use `interface{}` as a return type.** This is an anti-pattern because it makes it impossible to use code completion in an IDE. This is because the IDE can not know what the return type is.
- **Don't add more to the client class.** The client class is already too big and should not be made bigger. Instead, the client class should be split into multiple classes. Because something needs a class doesn't nessicarily mean it should be part of that class.
