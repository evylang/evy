# Learn

The learn directory contains all essential code and content for the Evy courses
on https://learn.evy.dev. This includes the markdown files, images, Evy code
samples, but also Firebase and Firestore utilities for answer verification
and progress tracking. However, frontend code is not included in this directory.

The target structure of Evy learn materials is, inspired by [Khanacademy]:

    Course -> Unit -> Exercise -> Question

We directly map this hierarchy to the file system:

    courses/
      fundamentals/                # course
        sequences/                 # unit
          print/                   # exercise
            print1.md          # question
            print2.md
            ...

Questions are encoded in markdown files. Each question has a YAML front matter section, e.g.:

     type: question
     difficulty: easy # "easy", "medium", "hard", "retriable"
     answer-type: single-choice # single-choice, multiple-choice, free-text, multiple-free-texts, program
     answer: c

[Khanacademy]: https://khanacademy.org

## `levy`

`levy` is a tool for creating and previewing Evy practice and learn materials.
It currently supports the following sub-commands:

- `export` the answer key to JSON
- `verify`
- `seal`
- `unseal`

Try it with

    make install
    levy export all --ignore-sealed pkg/learn/testdata/course1 out
    levy seal pkg/learn/testdata/course1/unit1/exercise1/question1.md

For sample error messages in case of failed verification, try

    levy verify pkg/learn/testdata/course1/unit1/err-exercise1/err-false-negative.md

To unseal sealed questions you need the secret private key. Here is a sample
usage with a private-key that is a test key. Do not use this key in
production!

    export EVY_LEARN_PRIVATE_KEY="MIIEpQIBAAKCAQEAuNEufiuryg/OZPKVUbaIRam1UNqju5binwrRzsOGWkM6DYKqxW2tA+O7dhg9do/Jm0lr+rkVqf8CR/HejD08n9OTsHe0NeblLwZncQX1J3ayyGsu+xAFxQ0hvFfG+Vy8KXJAgug6CCsaiVgBwOWPdfEOqEDv5S5XlnwQh9dxWB8m/1CTDmqSdIhYnzQQp13ZyumCRgrIHKSYPR3KCZD8KLRvkoIrF0DU18f6ASO7wjv7FBhgQ2ZAR/Yud/h6ceQKvAW0W3MmPiJblZhbrsPQGi7eZZo4K8aAvuzQmcYq17/E/e6MnOweoyik4lIAG0uGa7FiY5f9NVuir7JPA2lCLwIDAQABAoIBAQCJHEcNu4BbC5bnNUCpum0moVyue0X1KV8+9lvotQ27cRxkYYgnp9IvjIfKePlAODQtTC8bdqwnzdP3Y+zixZtwRxrOVEARrRZh6LJdGzpg6KKCJWJZR+2/3pokjEpFPRMq/GP3uikzXib1taC3ZpcjvI5PLL3MnLDGJ4xr+t1Pral8BXSILhUSQzgMFAB8+5V+zWnUPuPzCeym3VeYpZSdbSsR+CZnxy4vbB4cSj97M1MgBTOPocduE5cRrE8mumAk93dzBmKH+/potjLOMhCiJJFVtPO9GXLLduLAH9qKwSk2vJytIX8KwYFTCve2EKhMB9ydBhk09zVoELUle0mhAoGBAOc9agk5CahkNOVO1E3Cw1zK+Da+2LXYk2HhjTpOTkr2lKji8v1eSDkk5R72ZfPrI5s8sBrkW0OqPJVXDnmho78quWHTwxvJrnrIcuZLa1Kn4H+cHN81J9jGcim7kLPTZUcnU0RMR7Xn3lT61H5lB3LSFplRq52tqS5AaxaksS0tAoGBAMybQjceAVTihCHKFkaFV8Ys2dm5p5ejCzYklY+jA0UdTmHT6kmr13KIA6k61+s8kyZDaGutZ6lRyHuCfotL6j6jr8rsn/EbDikZ4/XhhO9+B+xJMXolKLFA+/pBPxNs7KLSjZ3mH7N0qzxbQzyVWF4BhSxTxIjWEGAtc1ZUJN5LAoGBAMYPFzRhE0GU2q2RkEwuRnDDNEiHvEw8/Td4HiPTkEGq4/ens2KKj6fKTyju+LIsM6oyF9BgyT6yoAN1tmM9rGf/qxr8av/xBa4K5EcWUA1S1vnV9/DCsad9iajvC2jK5tND/pDgGQfYWtlEoh7EX9Xb1hlqF2kNpnuEF3UkiNDdAoGBAJzuMFlKAEd0/VdVQsSQHYR4fhbKmMprWXwLj1L9+tIV6jqKaVZcIQFNZVF1OorIiSx94ydDdxCdE6H3sstwTJgCwCBqYTpyP+gyXXAHqwhtp/IJKZO/0HgzmZCWXqStlMpFqC0FhicEQxol/WoIOiDQFa6sCT/Sv/iko6QBIc4FAoGAMSC5SUsgUiHo6gvp2put1ySmJIVj3roqI6mAndi2hLVMalF1Q5F4X4HVHWqOj7QA7zpf3ATotCI4AbmfOwpFCZ4rEP0QsbV2uZ/3NhxwAE1MWrv+ht2ONe74sOYg7Z+XAjD7TW7We3KTewerVnC/VotKZ+3Eq2FgelSYDvlNmoQ="
    levy unseal pkg/learn/testdata/course1/unit1/exercise1/question1-sealed.md
