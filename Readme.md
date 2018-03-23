# Cosmos DB SQL upsert feature with Go lang

Currently we don't have an official Document DB SQL interface library. However, 

* [documentdb-go](https://github.com/a8m/documentdb-go) is one of the solution
* [documentdb-example](https://github.com/a8m/go-documentdb-example) is the sample

However, it doesn't support, the upsert feature. That is why I send [pull requrest](https://github.com/a8m/documentdb-go/pull/7) to the repo. 
I'm not sure if it is accepted or not. However 

This repo is the sample code for insert/update with upsert sample. 

You can use this feature with insall the library like this. (e.g. dep) 

```
[[constraint]]
  name = "github.com/TsuyoshiUshio/documentdb-go"
  branch = "feature/upsert"
```

# Limitation

Currently, I have some limitation for the current pull request. 

## 1. Can't avoid Upsert Status error

[This code]( https://github.com/a8m/documentdb-go/blob/master/client.go#L102-L106) validate the match of the status code. However, In case of Upsert, we have no way to know the status code in advance. 

It is not ideal, however, We can avoid to write the code like this. This should be also discuss in here. I'm not sure why the status matching is needed in here. 

```
		err := teamdb.Add(&v)
		errString := err.Error()
		if err != nil {
			// If you use upsert, the error happens with error sting ', '
			// which means status code validation error. Upsert can't predict which status code 200 or 201 in advance.
			if ", " != errString {
				fmt.Printf("upsert error! %d : '%s'", i, err.Error())
			}
		}
```
# 2 Partition Key is not supported

Partition Key is not supported (Should be the other pull request) It also require the new header and we need to pass the value to that header. maybe, refactor it to enable us to add custom header might be handy.

To implement Partition Key is not difficult just add instruction to the header. 
however, I need to discuss with the author.

* [Create Document](https://docs.microsoft.com/ja-jp/rest/api/documentdb/create-a-document)
* [Common Azure Cosmos DB REST request headers/x-ms-documentdb-partitionkey](https://docs.microsoft.com/ja-jp/rest/api/documentdb/common-documentdb-rest-request-headers)

# Resource 

* [Pull Request: Adding Upsert feature](https://github.com/a8m/documentdb-go/pull/7)
