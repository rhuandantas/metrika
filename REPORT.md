started running in sequence
identified problems in the persistence layer. the database would overload with lots of updates
 - update in batches
 - use caching to reduce load on the database
identified problem on the json writing:
 - use streaming json writers to handle large data sets
 - write json data in chunks instead of all at once
 - consider using a more efficient data format if json is too slow or resource-intensive
periodically save progress to avoid losing data in case of a crash
 - implement checkpointing to save the state of the process at regular intervals
 - avoid to conflict with other processes that might be writing to database
