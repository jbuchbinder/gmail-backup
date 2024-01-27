# GMAIL-BACKUP

Backup your entire gmail account.

## PREREQUISITES

1. You need to turn on IMAP for your gmail account, otherwise the utility will be unable to access your account. [Howto](https://support.google.com/a/answer/105694?hl=en)
2. If you are using 2FA (two factor authentication), you will need to create an [App Password](https://support.google.com/accounts/answer/185833?hl=en)

## USAGE

After compiling a binary, run like this:

```
gmail-backup -u USER@DOMAIN.COM -p YOURPASSWORDORAPPPASSWORD -d DESTDIR
```

or, create a .env file like:

```
USERNAME=me@my.net
PASSWORD=blahblahblahblah
```

and run it like this:

```
gmail-backup -d DESTDIR
```

