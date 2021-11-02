# Changes

## [1.24.0](https://www.github.com/googleapis/google-cloud-go/compare/bigquery/v1.23.0...bigquery/v1.24.0) (2021-09-27)


### Features

* **bigquery/migration:** Add PAUSED state to Subtask and add task details protos ([bddab08](https://www.github.com/googleapis/google-cloud-go/commit/bddab08dfd0b9a0a79b113a46a0dd84dba1f3d3b))


### Bug Fixes

* **bigquery/storage:** add missing read api retry setting on SplitReadStream ([797a9bd](https://www.github.com/googleapis/google-cloud-go/commit/797a9bdcb68c0c3ff7eef04cd3a3a0747937975b))

## [1.23.0](https://www.github.com/googleapis/google-cloud-go/compare/bigquery/v1.22.0...bigquery/v1.23.0) (2021-09-23)


### Features

* **bigquery/reservation:**
  * Deprecated SearchAssignments in favor of SearchAllAssignments
  * feat: Reservation objects now contain a creation time and an update time
  * feat: Added commitment_start_time to capacity commitments
  * feat: Force deleting capacity commitments is allowed while reservations with active assignments exist
  * feat: ML_EXTERNAL job type is supported
  * feat: Optional id can be passed into CreateCapacityCommitment and CreateAssignment
  * docs: Clarified docs for None assignments
  * fix!: Fixed pattern for BiReservation object BREAKING_CHANGE: Changed from `bireservation` to `biReservation`
  * ([d9ce9d0](https://www.github.com/googleapis/google-cloud-go/commit/d9ce9d0ee64f59c4e07ce4752bfd721051a95ac7))
* **bigquery/storage/managedwriter:** BREAKING CHANGE: changeAppendRows behavior ([#4729](https://github.com/googleapis/google-cloud-go/pull/4729))
* **bigquery/storage:** add BigQuery Storage Write API v1 ([e52c204](https://www.github.com/googleapis/google-cloud-go/commit/e52c2042a2b7cdd7dd799a561421f32fecc5d1d2))
* **bigquery/storage:** migrate managedwriter to v1 write from v1beta2 ([#4788](https://github.com/googleapis/google-cloud-go/pull/4788))
* **bigquery:** add session and connection support ([#4754](https://www.github.com/googleapis/google-cloud-go/issues/4754)) ([e846dfd](https://www.github.com/googleapis/google-cloud-go/commit/e846dfdefbba88320088667525e5fdd966c80c4b))
* **bigquery:** expose the query source of a rowiterator via SourceJob() ([#4748](https://github.com/googleapis/google-cloud-go/pull/4748))

## [1.22.0](https://www.github.com/googleapis/google-cloud-go/compare/bigquery/v1.21.0...bigquery/v1.22.0) (2021-08-30)


### Features

* **bigquery/storage/managedwriter/adapt:** add NormalizeDescriptor ([#4681](https://www.github.com/googleapis/google-cloud-go/issues/4681)) ([c54aa74](https://www.github.com/googleapis/google-cloud-go/commit/c54aa74f7a0574cbbe3f65dc90b96cf5a0b1aa88))
* **bigquery/storage/managedwriter:** more metrics instrumentation ([#4690](https://www.github.com/googleapis/google-cloud-go/issues/4690)) ([9505384](https://www.github.com/googleapis/google-cloud-go/commit/9505384b2c771d7d0c95f7786744bdf76174c706))

## [1.21.0](https://www.github.com/googleapis/google-cloud-go/compare/bigquery/v1.20.1...bigquery/v1.21.0) (2021-08-16)


### Features

* **bigquery/storage/managedwriter:** add project autodetection ([#4605](https://www.github.com/googleapis/google-cloud-go/issues/4605)) ([d8cc9be](https://www.github.com/googleapis/google-cloud-go/commit/d8cc9be6f0314f585f708638834abfc209799724))
* **bigquery/storage/managedwriter:** improve protobuf support ([#4589](https://www.github.com/googleapis/google-cloud-go/issues/4589)) ([a455082](https://www.github.com/googleapis/google-cloud-go/commit/a45508272a730e0ad81021695d2d8564e7c81631))
* **bigquery/storage/managedwriter:** more instrumentation support ([#4601](https://www.github.com/googleapis/google-cloud-go/issues/4601)) ([ff488c8](https://www.github.com/googleapis/google-cloud-go/commit/ff488c86b9c1a1f02397bb579905fa049e59ac05))
* **bigquery:** switch to centralized project autodetect logic ([#4625](https://www.github.com/googleapis/google-cloud-go/issues/4625)) ([18ff070](https://www.github.com/googleapis/google-cloud-go/commit/18ff070b8baa3ed7d324ca9ea00dcd66d7742340))


### Bug Fixes

* **bigquery/storage/managedwriter:** support non-default regions ([#4566](https://www.github.com/googleapis/google-cloud-go/issues/4566)) ([68418f9](https://www.github.com/googleapis/google-cloud-go/commit/68418f9e340def179eb5556aea433c0d07000b79))

### [1.20.1](https://www.github.com/googleapis/google-cloud-go/compare/bigquery/v1.20.0...bigquery/v1.20.1) (2021-08-06)


### Bug Fixes

* **bigquery/storage/managedwriter:** fix flowcontroller double-release ([#4555](https://www.github.com/googleapis/google-cloud-go/issues/4555)) ([67facd9](https://www.github.com/googleapis/google-cloud-go/commit/67facd9697e931e193f3cd8e188f1dd819ba31eb))

## [1.20.0](https://www.github.com/googleapis/google-cloud-go/compare/bigquery/v1.19.0...bigquery/v1.20.0) (2021-07-30)


### Features

* **bigquery/connection:** add cloud spanner connection support ([458f15b](https://www.github.com/googleapis/google-cloud-go/commit/458f15bb6f1193ce83dbfc7a82c3f2a672f52c06))
* **bigquery/storage/managedwriter/adapt:** add schema -> proto support ([#4375](https://www.github.com/googleapis/google-cloud-go/issues/4375)) ([4ff6243](https://www.github.com/googleapis/google-cloud-go/commit/4ff62433f58c1c92976a66e890b7d5394198f77b))
* **bigquery/storage/managedwriter:** add append stream plumbing ([#4452](https://www.github.com/googleapis/google-cloud-go/issues/4452)) ([b085384](https://www.github.com/googleapis/google-cloud-go/commit/b0853846a34a32ca45deb92a3cc6ab843473acd8))
* **bigquery/storage/managedwriter:** add base client ([#4422](https://www.github.com/googleapis/google-cloud-go/issues/4422)) ([4f7193b](https://www.github.com/googleapis/google-cloud-go/commit/4f7193b74f4b1954cf7b664d61b5cc9805337e84))
* **bigquery/storage/managedwriter:** add flow controller ([#4404](https://www.github.com/googleapis/google-cloud-go/issues/4404)) ([9dc78e0](https://www.github.com/googleapis/google-cloud-go/commit/9dc78e073b5f69037c6328460554c4354fcee11f))
* **bigquery/storage/managedwriter:** add opencensus instrumentation ([#4512](https://www.github.com/googleapis/google-cloud-go/issues/4512)) ([73b6f5e](https://www.github.com/googleapis/google-cloud-go/commit/73b6f5e012d0b89d36850cb986fd7e288bf1e3c5))
* **bigquery/storage/managedwriter:** add state tracking ([#4407](https://www.github.com/googleapis/google-cloud-go/issues/4407)) ([4638e17](https://www.github.com/googleapis/google-cloud-go/commit/4638e17dacd1fa76f9976f44974c4037fe4358dc))
* **bigquery/storage/managedwriter:** naming and doc improvements ([#4508](https://www.github.com/googleapis/google-cloud-go/issues/4508)) ([663c899](https://www.github.com/googleapis/google-cloud-go/commit/663c899c3b8aa751527d24f541d964f2ba91a233))
* **bigquery/storage/managedwriter:** wire in flow controller ([#4501](https://www.github.com/googleapis/google-cloud-go/issues/4501)) ([40571fa](https://www.github.com/googleapis/google-cloud-go/commit/40571fa0e3b5ab326fd592a6907061c2304893aa))
* **bigquery:** add more dml statistics to query statistics ([#4405](https://www.github.com/googleapis/google-cloud-go/issues/4405)) ([99d5728](https://www.github.com/googleapis/google-cloud-go/commit/99d57282f6668de91390ad29a888a89209689f39))
* **bigquery:** support decimalTargetType prioritization ([#4343](https://www.github.com/googleapis/google-cloud-go/issues/4343)) ([95a27f7](https://www.github.com/googleapis/google-cloud-go/commit/95a27f711a1c7dfdaa16ae5d3c52644769b6fc39))
* **bigquery:** support multistatement transaction statistics in jobs ([#4485](https://www.github.com/googleapis/google-cloud-go/issues/4485)) ([4565eb7](https://www.github.com/googleapis/google-cloud-go/commit/4565eb7fe730eade294fb3baa85bd255df008bfa))


### Bug Fixes

* **bigquery/storage/managedwriter:** fix double-close error, add tests ([#4502](https://www.github.com/googleapis/google-cloud-go/issues/4502)) ([c6cf659](https://www.github.com/googleapis/google-cloud-go/commit/c6cf6590a41368885b7399c993c47dc965862558))

## [1.19.0](https://www.github.com/googleapis/google-cloud-go/compare/bigquery/v1.18.0...bigquery/v1.19.0) (2021-06-29)


### Features

* **bigquery/storage:** Add ZSTD compression as an option for Arrow. ([770db30](https://www.github.com/googleapis/google-cloud-go/commit/770db3083270d485d265362fe5a4b2a1b23619ff))
* **bigquery/storage:** remove alpha client ([#4100](https://www.github.com/googleapis/google-cloud-go/issues/4100)) ([a2d137d](https://www.github.com/googleapis/google-cloud-go/commit/a2d137d233e7a401976fbe1fd8ff81145dda515d)), refs [#4098](https://www.github.com/googleapis/google-cloud-go/issues/4098)
* **bigquery:** add support for parameterized types ([#4103](https://www.github.com/googleapis/google-cloud-go/issues/4103)) ([a2330e4](https://www.github.com/googleapis/google-cloud-go/commit/a2330e4d66c0a1832fb3b9e23a33c006c9345c28))
* **bigquery:** add support for snapshot/restore ([#4112](https://www.github.com/googleapis/google-cloud-go/issues/4112)) ([4c12b42](https://www.github.com/googleapis/google-cloud-go/commit/4c12b424eec06c7d87244eaa922995bbe6e46e7e))
* **bigquery:** add support for user defined TVF ([#4043](https://www.github.com/googleapis/google-cloud-go/issues/4043)) ([37607b4](https://www.github.com/googleapis/google-cloud-go/commit/37607b4afbc4c42baa4a931a9a86cddcc6d885ca))
* **bigquery:** enable project autodetection, expose project ids further ([#4312](https://www.github.com/googleapis/google-cloud-go/issues/4312)) ([267787e](https://www.github.com/googleapis/google-cloud-go/commit/267787eb245d9307cf78304c1ce34bdfb2aaf5ab))
* **bigquery:** support job deletion ([#3935](https://www.github.com/googleapis/google-cloud-go/issues/3935)) ([363ba03](https://www.github.com/googleapis/google-cloud-go/commit/363ba03e1c3c813749a65ff3c050877ce4f60016))
* **bigquery:** support nullable params and geography params ([#4225](https://www.github.com/googleapis/google-cloud-go/issues/4225)) ([43755d3](https://www.github.com/googleapis/google-cloud-go/commit/43755d38b5d928222127cc6be26183d6bfbb1cb4))


### Bug Fixes

* **bigquery:** minor rename to feature that's not yet in a release ([#4320](https://www.github.com/googleapis/google-cloud-go/issues/4320)) ([ef8d138](https://www.github.com/googleapis/google-cloud-go/commit/ef8d1386149cff28ae6258ab167789bae6af6407))
* **bigquery:** update streaming insert error test ([#4321](https://www.github.com/googleapis/google-cloud-go/issues/4321)) ([12f3042](https://www.github.com/googleapis/google-cloud-go/commit/12f3042716d51fb0d7a23071d00a20f9751bac91))

## [1.18.0](https://www.github.com/googleapis/google-cloud-go/compare/bigquery/v1.17.0...bigquery/v1.18.0) (2021-05-06)


### Features

* **bigquery/storage:** new JSON type through BigQuery Write ([9029071](https://www.github.com/googleapis/google-cloud-go/commit/90290710158cf63de918c2d790df48f55a23adc5))
* **bigquery:** augment retry predicate to support additional errors ([#4046](https://www.github.com/googleapis/google-cloud-go/issues/4046)) ([d4af6f7](https://www.github.com/googleapis/google-cloud-go/commit/d4af6f7707b3c5ee12cde53c7485a9b743034119))
* **bigquery:** expose ParquetOptions for loads and external tables ([#4016](https://www.github.com/googleapis/google-cloud-go/issues/4016)) ([f9c4ccb](https://www.github.com/googleapis/google-cloud-go/commit/f9c4ccb6efb271c421edf3f67d5249b1cfb0ecb2))
* **bigquery:** support mutable clustering configuration ([#3950](https://www.github.com/googleapis/google-cloud-go/issues/3950)) ([0ab30da](https://www.github.com/googleapis/google-cloud-go/commit/0ab30dadc43ae85354dc12a4130ecfcc56273882))

## [1.17.0](https://www.github.com/googleapis/google-cloud-go/compare/bigquery/v1.15.0...bigquery/v1.17.0) (2021-04-08)


### Features

* **bigquery/storage:** add a Arrow compression options (Only LZ4 for now). feat: Return schema on first ReadRowsResponse. doc: clarify limit on filter string. ([2b02a03](https://www.github.com/googleapis/google-cloud-go/commit/2b02a03ff9f78884da5a8e7b64a336014c61bde7))
* **bigquery/storage:** deprecate bigquery storage v1alpha2 API ([9cc6d2c](https://www.github.com/googleapis/google-cloud-go/commit/9cc6d2cce96235b0a144c1c6b48eff496f9e5fa7))
* **bigquery/storage:** updates for v1beta2 storage API - Updated comments on BatchCommitWriteStreams - Added new support Bigquery types BIGNUMERIC and INTERVAL to TableSchema - Added read rows schema in ReadRowsResponse - Misc comment updates ([48b4e59](https://www.github.com/googleapis/google-cloud-go/commit/48b4e596206cef879194d2888186d603a6f51292))
* **bigquery:** export HivePartitioningOptions in load job configurations ([#3877](https://www.github.com/googleapis/google-cloud-go/issues/3877)) ([7c759be](https://www.github.com/googleapis/google-cloud-go/commit/7c759be074ce1f6b8ccce88c86dbe49bd38fd6b5))
* **bigquery:** support type alias names for numeric/bignumeric schemas. ([#3760](https://www.github.com/googleapis/google-cloud-go/issues/3760)) ([2ee6bf4](https://www.github.com/googleapis/google-cloud-go/commit/2ee6bf451524fc1f9735634320a55ca0b07d3d8b))

## v1.16.0

- Updates to various dependencies.

## [1.15.0](https://www.github.com/googleapis/google-cloud-go/compare/bigquery/v1.14.0...v1.15.0) (2021-01-14)


### Features

* **bigquery:** add reservation usage stats to query statistics ([#3403](https://www.github.com/googleapis/google-cloud-go/issues/3403)) ([112bcde](https://www.github.com/googleapis/google-cloud-go/commit/112bcdeb7cee1b44f337d3e5398a0d0820e93162))
* **bigquery:** add support for allowing Javascript UDFs to indicate determinism ([#3534](https://www.github.com/googleapis/google-cloud-go/issues/3534)) ([2f417a3](https://www.github.com/googleapis/google-cloud-go/commit/2f417a39d93402fbb1e5e3001645019782d7d656)), refs [#3533](https://www.github.com/googleapis/google-cloud-go/issues/3533)


### Bug Fixes

* **bigquery:** address possible panic due to offset checking in handleInsertErrors  ([#3524](https://www.github.com/googleapis/google-cloud-go/issues/3524)) ([5288511](https://www.github.com/googleapis/google-cloud-go/commit/52885115af3e95cdfd1ec784837fb1df7fe01446)), refs [#3519](https://www.github.com/googleapis/google-cloud-go/issues/3519)

## [1.14.0](https://www.github.com/googleapis/google-cloud-go/compare/bigquery/v1.13.0...v1.14.0) (2020-12-04)


### Features

* **bigquery:** add support for bignumeric ([#2779](https://www.github.com/googleapis/google-cloud-go/issues/2779)) ([ea3cde5](https://www.github.com/googleapis/google-cloud-go/commit/ea3cde55ad3d8d843bce8d023747cf69552850b5))
* **bigquery:** expose hive partitioning options ([#3240](https://www.github.com/googleapis/google-cloud-go/issues/3240)) ([fa77efa](https://www.github.com/googleapis/google-cloud-go/commit/fa77efa1a1880ff89307d54cc7e9e8c09430e4e2))

## v1.13.0

* Support retries for specific http2 transport race.
* Remove unused datasource client from bigquery/datatransfer.
* Adds support for authorized User Defined Functions (UDFs).
* Documentation improvements.
* Various updates to autogenerated clients.


## v1.12.0

* Adds additional retry support for table deletion.
* Various updates to autogenerated clients.

## v1.11.2

* Addresses issue with consuming query results using an iterator.Pager

## v1.11.1

* Addresses issue with optimized query path changes, released
  in v1.11.0

## v1.11.0

* Add support for optimized query path.
* Documentation improvements.
* Fix issue related to the ReturnType of a bigquery Routine.
* Various updates to autogenerated clients.

## v1.10.0

* Support for Infinity/-Infinity/NaN values in NullFloat64.
* Updates to RowIterator to address issues related to retrieving query
  results without explicit destination table references.
* Various updates to autogenerated clients.

## v1.9.0

* SchemaFromJSON will now accept alias type names (e.g. INT64 vs INTEGER, STRUCT vs RECORD).
* Support for IAM on table resources.
* Various updates to autogenerated clients.

## v1.8.0

* Add support for hourly time partitioning.
* Various updates to autogenerated clients.

## v1.7.0

* Add support for extracting BQML models to cloud storage.
* Add support for specifying projected fields when ingesting
  datastore backups.
* Fix issue related to defining a range partitioning range
  using default values.
* Add bigquery/reservation/v1 API.
* Various updates to autogenerated clients.

## v1.6.0

* Add support for materialized views.
* Add support for policy tags (column ACLs).
* Add bigquery/connection/v1beta1 API.
* Documentation improvements.
* Various updates to autogenerated clients.

## v1.5.0

* Add v1 endpoint for bigquerystorage API.
* Improved error message in bigquery.PutMultiError.
* Various updates to autogenerated clients.

## v1.4.0

* Add v1beta2, v1alpha2 endpoints for bigquerystorage API.

* Location is now reported as part of TableMetadata.

## v1.3.0

* Add Description field for Routine entities.

* Add support for iamMember entities on dataset ACLs.

* Address issue when constructing a Pager from a RowIterator
  that referenced a result with zero result rows.

* Add support for integer range partitioning, which affects
  table creation directly and via query/load jobs.

* Add opt-out support for streaming inserts via experimental
  `NoDedupeID` sentinel.

## v1.2.0

* Adds support for scripting feature, which includes script statistics
  and the ability to list jobs run as part of a script query.

* Updates default endpoint for BigQuery from www.googleapis.com
  to bigquery.googleapis.com.

## v1.1.0

* Added support for specifying default `EncryptionConfig` settings on the
  dataset.

* Added support for `EncyptionConfig` as part of an ML model.

* Added `Relax()` to make all fields within a `Schema` nullable.

* Added a `UseAvroLogicalTypes` option when defining an avro extract job.

## v1.0.1

This patch release is a small fix to the go.mod to point to the post-carve out
cloud.google.com/go.

## v1.0.0

This is the first tag to carve out bigquery as its own module. See:
https://github.com/golang/go/wiki/Modules#is-it-possible-to-add-a-module-to-a-multi-module-repository.
