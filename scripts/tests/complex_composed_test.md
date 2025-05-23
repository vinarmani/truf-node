### Deploy & Init

This will be our calculated table to see if it works as we want:
| date       | p1   | p2   | p3   |
| ---------- | ---- | ---- | ---- |
| 2021-01-01 |      |      | 3    |
| 2021-01-02 | 4    | 5    | 6    |
| 2021-01-03 |      |      | 9    |
| 2021-01-04 | 10   |      |      |
| 2021-01-05 | 13   |      | 15   |
| 2021-01-06 |      | 17   | 18   |
| 2021-01-07 | 19   | 20   |      |
| 2021-01-08 |      | 23   |      |
| 2021-01-09 | 25   |      |      |
| 2021-01-10 |      |      | 30   |
| 2021-01-11 |      | 32   |      |
| 2021-01-12 |      |      |      |
| 2021-01-13 |      |      | 39   |

- Missing values that have past data for the same primitive stream will be filled forward
- Missing values that do not have past values should be DISCONSIDERED, not contributing to the weighted average. It means its weight should be 0.

Deploy contract and setup weights of 1, 2, 3 for p1, p2, p3 respectively. We set different weights to see if the calculation is correct.
```shell
../../.build/kwil-cli database drop complex_composed_a --sync
../../.build/kwil-cli database deploy -p=../../internal/contracts/composed_stream_template.kf --name complex_composed_a --sync
../../.build/kwil-cli database execute --action=init -n=complex_composed_a --sync 
my_wallet=$(../../.build/kwil-cli account id)
../../.build/kwil-cli database execute data_providers:$my_wallet,$my_wallet,$my_wallet stream_ids:p1,p2,p3 weights:1,2,3 --action=set_taxonomy -n=complex_composed_a --sync
```

Deploy primitives and insert records
```shell
# drop all primitive streams
../../.build/kwil-cli database drop p1 
../../.build/kwil-cli database drop p2 
../../.build/kwil-cli database drop p3 --sync

# deploy primitive streams
../../.build/kwil-cli database deploy -p=../../internal/contracts/primitive_stream_template.kf --name=p1
../../.build/kwil-cli database deploy -p=../../internal/contracts/primitive_stream_template.kf --name=p2
../../.build/kwil-cli database deploy -p=../../internal/contracts/primitive_stream_template.kf --name=p3 --sync

# init primitive streams
../../.build/kwil-cli database execute --action=init -n=p1
../../.build/kwil-cli database execute --action=init -n=p2
../../.build/kwil-cli database execute --action=init -n=p3 --sync

# date 2021-01-01
## no data for p1
## no data for p2
../../.build/kwil-cli database execute --action=insert_record -n=p3 date_value:2021-01-01 value:3

# date 2021-01-02
../../.build/kwil-cli database execute --action=insert_record -n=p1 date_value:2021-01-02 value:4
../../.build/kwil-cli database execute --action=insert_record -n=p2 date_value:2021-01-02 value:5
../../.build/kwil-cli database execute --action=insert_record -n=p3 date_value:2021-01-02 value:6

# date 2021-01-03
## no data for p1
## no data for p2
../../.build/kwil-cli database execute --action=insert_record -n=p3 date_value:2021-01-03 value:9

# date 2021-01-04
../../.build/kwil-cli database execute --action=insert_record -n=p1 date_value:2021-01-04 value:10
## no data for p2
## no data for p3

# date 2021-01-05
../../.build/kwil-cli database execute --action=insert_record -n=p1 date_value:2021-01-05 value:13
## no data for p2
../../.build/kwil-cli database execute --action=insert_record -n=p3 date_value:2021-01-05 value:15

# date 2021-01-06
## no data for p1
../../.build/kwil-cli database execute --action=insert_record -n=p2 date_value:2021-01-06 value:17
../../.build/kwil-cli database execute --action=insert_record -n=p3 date_value:2021-01-06 value:18

# date 2021-01-07
../../.build/kwil-cli database execute --action=insert_record -n=p1 date_value:2021-01-07 value:19
../../.build/kwil-cli database execute --action=insert_record -n=p2 date_value:2021-01-07 value:20
## no data for p3

# date 2021-01-08
## no data for p1
../../.build/kwil-cli database execute --action=insert_record -n=p2 date_value:2021-01-08 value:23
## no data for p3

# date 2021-01-09
../../.build/kwil-cli database execute --action=insert_record -n=p1 date_value:2021-01-09 value:25
## no data for p2
## no data for p3

# date 2021-01-10
## no data for p1
## no data for p2
../../.build/kwil-cli database execute --action=insert_record -n=p3 date_value:2021-01-10 value:30 --sync

# date 2021-01-11
## no data for p1
../../.build/kwil-cli database execute --action=insert_record -n=p2 date_value:2021-01-11 value:32
## no data for p3

# date 2021-01-12
## no data for p1
## no data for p2
## no data for p3

# date 2021-01-13
## no data for p1
## no data for p2
../../.build/kwil-cli database execute --action=insert_record -n=p3 date_value:2021-01-13 value:39 --sync
```

# get record for each primitive stream
```shell
../../.build/kwil-cli database call --action=get_record date_from:2021-01-01 date_to:2021-01-10 -n=p1
../../.build/kwil-cli database call --action=get_record date_from:2021-01-01 date_to:2021-01-10 -n=p2
../../.build/kwil-cli database call --action=get_record date_from:2021-01-01 date_to:2021-01-10 -n=p3
```


```shell
../../.build/kwil-cli database call --action=get_record date_from:2021-01-01 date_to:2021-01-13 -n=complex_composed_a
```

This is the expected result, calculated from an spreadsheet:

| Date       | Result |
|------------|--------|
| 2021-01-01 | 3      |
| 2021-01-02 | 5.333  |
| 2021-01-03 | 6.833  |
| 2021-01-04 | 7.833  |
| 2021-01-05 | 11.333 |
| 2021-01-06 | 16.833 |
| 2021-01-07 | 18.833 |
| 2021-01-08 | 19.833 |
| 2021-01-09 | 20.833 |
| 2021-01-10 | 26.833 |
| 2021-01-11 | 29.833 | 
| 2021-01-13 | 34.333 |

Note the missing value on 2021-01-12, because there's no data point in any primitive stream.

# get latest value
```shell
../../.build/kwil-cli database call --action=get_record -n=complex_composed_a
```

This is the expected result:

| Date       | Result |
|------------|--------|
| 2021-01-13 | 34.333 |

# get from an empty date in the middle
```shell
../../.build/kwil-cli database call --action=get_record date_from:2021-01-12 date_to:2021-01-12 -n=complex_composed_a
```

This is the expected result:

| Date       | Result |
|------------|--------|
| 2021-01-11 | 29.833 |

Note that the result is the same as the previous date, because the missing value on 2021-01-12, so the most recent value is 2021-01-11.

Let's also check the index of the latest value:
```shell
../../.build/kwil-cli database call --action=get_index -n=complex_composed_a
```

Expected result:

| Date       | Result  |
|------------|---------|
| 2021-01-13 | 967.500 |

# Check index for all dates
```shell
../../.build/kwil-cli database call --action=get_index date_from:2021-01-01 date_to:2021-01-13 -n=complex_composed_a
```

Expected result:

| Date       | Result  |
|------------|---------|
| 2021-01-01 | 100.000 |
| 2021-01-02 | 150.000 |
| 2021-01-03 | 200.000 |
| 2021-01-04 | 225.000 |
| 2021-01-05 | 337.500 |
| 2021-01-06 | 467.500 |
| 2021-01-07 | 512.500 |
| 2021-01-08 | 532.500 |
| 2021-01-09 | 557.500 |
| 2021-01-10 | 757.500 |
| 2021-01-11 | 817.500 |
| 2021-01-13 | 967.500 |

Note that the final index is 967.5 is NOT the same as dividing the record for the last day by the first record, because the weights are applied on the index percentages themselves. See https://system.docs.truflation.com/backend/cpi-calculations/workflow/aggregated-indexes for more information.
