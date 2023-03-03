# map2xml
Golang map2xml is a wrapper around encoding/xml to enable it to marshal interface maps.

It is based upon a builder pattern to enable developers to execute simple instructions without having to consider features they don't want to use.

## Installation
```
go get github.com/yoda-of-soda/map2xml
```
## How to use it
You start with the ```map2xml.New()``` function and insert a ```map[string]interface{}```, e.g. 

```golang
inputMap := map[string]interface{}{
                "first_name": "No",
                "last_name":  "Name",
                "age":        42,
                "got_a_job":  true,
                "address": map[string]interface{}{
                    "street":   "124 Oxford Street",
                    "city":     "Somewhere",
                    "zip_code": 32784,
                    "state":    "Deep state",
                },
            }
```

From here you can set up various properties of how your map will be converted to XML. The current options are:
* Root name
* Root attributes
* Indentation (like MarshalIndent)
* CData (whether all fields will be wrapped with cdata tags or not)

When the configuration has been set up, you can print the configuration with the ```Print()``` method or you can marshal the map into xml bytes using ```Marshal()``` or directly to an xml string using ```MarshalToString()```. 


You can split it up into configuration and execution like this:
```golang
config := map2xml.New(inputMap)
config.WithIndent("", "  ")
config.WithRoot("person", map[string]string{"mood": "happy"})
xmlBytes, err := config.Marshal()
```

All these functions can be put together in a single line like this:
```golang
xmlString, err := map2xml.New(inputMap).AsCData().WithIndent("", "  ").WithRoot("person", map[string]string{"mood": "happy"}).MarshalToString()
```

You can even put the ```Print()``` function in the loop, e.g.
```golang
xmlString, err := map2xml.New(inputMap, "person").WithIndent("", "  ").Print().MarshalToString()
```


It is very important to know that either ```Marshal()``` or ```MarshalToString()``` should be the last function to be called.
