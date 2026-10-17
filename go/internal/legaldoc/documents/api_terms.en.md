## About the Establishment of These Terms

We established these Terms on October 15, 2026, and they take effect on November 15, 2026.
Most of those who have been using the API can continue to use it without changing how they use it.
Please check by the effective date that your use complies with these Terms.
If you do not agree to these Terms, stop using the API by the effective date and delete the applications and access tokens you registered.

These Terms set out the following points:

- Using the API is free of charge, and you can use it without individual permission, including for commercial purposes
- We distinguished displaying user information within an application from republishing it outside Annict
- We set out the conditions for storing and caching retrieved data, and prohibited redistribution
- We set out the conditions for sharing one access token in applications that do not require an Annict account
- We set out the conditions for retrieving work information not provided by the API from the website by automated means
- We set out the conditions for use in AI and machine learning, and for handling images
- We set out rate limits, the transfer of OAuth applications, and the display of the Annict name and logo
- When we stop providing the API, we will announce it at least one year in advance

Purposes that the Operator individually permitted before these Terms took effect remain permitted after they take effect ("Supplementary Provisions").
These Terms are newly established, and there is no previous version.

---

The operator (the "Operator") of the anime tracking service "Annict" (the "Service") sets out the conditions for using the API of the Service as these API Terms of Use (these "Terms") as follows.
These Terms apply in addition to the [Terms of Service](/en/terms).

This English version is a translation of the Japanese original.
If there is any discrepancy in interpretation between the Japanese version and the English version, the Japanese version prevails.
These Terms are governed by the laws of Japan.

## Article 1 (Scope)

These Terms apply to everyone ("Developers") who uses any of the following (collectively, the "API"):

- REST API V1
- GraphQL API
- Application authorization through OAuth
- Personal access tokens
- APIs we provide in the future (including REST API V2)

These Terms also apply when you retrieve information from the website of the Service by automated means ("Article 8 (Retrieval from the Website by Automated Means)").

Matters not set out in these Terms are governed by the Terms of Service.
If these Terms and the Terms of Service differ regarding the use of the API, these Terms prevail.

## Article 2 (Definitions)

The terms used in these Terms have the following meanings:

- Work information: information other than information about users of the Service, such as works, episodes, broadcast schedules, characters, people, organizations, casts, staff, and series
- User information: information about users of the Service, such as profiles, records, reviews, watching statuses, libraries, and follows
- Application: an app, website, service, or program that a Developer builds using the API

## Article 3 (Fees and Commercial Use)

Using the API is free of charge.
We do not offer paid agreements either.

As long as Developers comply with these Terms, they can use the API without individual permission, whether for commercial or non-commercial purposes.
This includes use in applications that earn revenue through advertising or affiliate programs, and in applications with paid features.

Individual permission from the Operator is required only in the following cases:

- Using the API for AI and machine learning for commercial purposes ("Article 9 (Use in AI and Machine Learning)")
- Providing retrieved data to third parties as a dataset ("Article 6 (Handling of Retrieved Data)")

## Article 4 (Authentication)

The Operator recommends authorizing applications through OAuth.
With OAuth, an application can request only the permissions (scopes) it needs, and users can revoke the authorization for each application.

When you use a personal access token, comply with the following:

- Store the access token so that third parties do not learn it
- In applications that ask users to enter a personal access token, explain to users what the token can do and that they should not share the token with others
- Do not give your access token to third parties
  However, this does not apply to the case in "Article 5 (Shared Tokens Embedded in Distributed Applications)"

## Article 5 (Shared Tokens Embedded in Distributed Applications)

In applications that do not require users to create an Annict account, a Developer may embed one of their own personal access tokens in the application and use it for the requests of all users.
In that case, comply with all of the following conditions:

- Use the token to retrieve only work information, and do not retrieve user information
- Use a token issued with only the read (`read`) scope
- Understand that the Operator may revoke the token, and build the application so that the token can be replaced

A token embedded in an application may be extracted by a third party.
If an extracted token is misused, the Operator may revoke the token without contacting the Developer in advance.

## Article 6 (Handling of Retrieved Data)

Developers must not provide data retrieved through the API as is, or in bulk, in a form that third parties can extract.
This includes distributing it as files, publishing it as a dataset, and re-providing it through your own API.
This does not apply if the Operator individually permits it.
Displaying work information on the screens or pages of an application does not fall under this.

User information is handled separately for displaying it within an application and for republishing it outside Annict.

- Display: showing user information to users of the application as a feature of Annict
  This includes showing usernames and reviews in a client app
  You may display user information as long as you comply with all of the following conditions
  - Show the content retrieved from the API without changing it
  - Do not show it beyond the scope made public on the Service
  - Stop showing information that has been deleted or made private on the Service
  - Do not accumulate copies beyond what is necessary for display
- Republishing: putting user information in public materials outside Annict
  This includes posting it on websites, quoting it in papers, including it in datasets, and posting it on social media
  Do not republish user information unless the user to whom it belongs has consented

Results of statistics or aggregation that do not identify individuals are not republishing.
You can publish them freely.

## Article 7 (Storing and Caching Data)

Developers may store and cache data retrieved through the API to the extent necessary to provide the application.
It does not matter whether the data is stored on the Developer's servers or on users' devices.

Work information may be handled as follows:

- There is no limit on the storage period or the number of items
  You may store all of it
- You do not need to erase it after you stop using the API
- You may keep the IDs of works and other items to identify works or link them within the application

Work information may be corrected, merged, deleted, or made private on the Service.
We would appreciate it if you retrieve it again periodically and reflect it in your application.
People whose information has been delisted can no longer be found through the API when you retrieve the information again.
In that case, remove that person's information from your application as well.
Guidelines on how often to retrieve information again are listed in "Annict Rate Limits" in the [developer help](https://wikino.app/s/annict/pages/357) (in Japanese).

Delete user information in the following cases.
Erase anything that remains in backups once the normal retention period has passed.

- When a user revokes the authorization of the application: that user's user information
- When you stop using the API: all user information

## Article 8 (Retrieval from the Website by Automated Means)

Retrieve information provided by the API through the API.

You may retrieve work information that is not provided by the API (such as links to streaming services) from the website of the Service by automated means without individual permission, as long as you comply with the following:

- Follow robots.txt
- Retrieve it within the rate limits in "Article 11 (Rate Limits)"

You must not retrieve user information from the website by automated means.
This is because the Terms of Service prohibit it.

The HTML structure of the website may change without notice.
The Operator does not guarantee that retrieval from the website by automated means will keep working, and does not provide support for it.
If the Operator asks you to stop retrieving, stop.

When retrieving by automated means, we would appreciate it if you include the name of the application and contact information (a website URL or an email address) in the `User-Agent` header.
When the load becomes high, we may be able to contact you before blocking access.

## Article 9 (Use in AI and Machine Learning)

You may use data retrieved through the API for AI and machine learning.
However, the following conditions apply depending on the purpose:

- Research or non-commercial purposes: you may use the data without individual permission as long as you comply with the following conditions
  - Do not redistribute the retrieved data ("Article 6 (Handling of Retrieved Data)")
  - Do not include user information in deliverables, such as papers or example outputs of models, in a form that identifies individuals
- Commercial purposes: individual permission from the Operator is required

For either purpose, retrieve the data through the API.
The Operator does not provide copies of the database (data dumps).

## Article 10 (Handling of Images)

The copyrights in the images of works provided by the API and the images displayed on the website belong to their respective rights holders.
The images of works provided by the API include URLs such as images from the official websites of works and profile images of official social media accounts.
The Operator is not in a position to license the use of these images.

Use images under your own responsibility, to the extent permitted by laws and regulations (such as displaying them alongside links that introduce works).
You must not do the following:

- Redistribute images
- Modify images
- Provide images as a dataset

Image URLs include broken links.
The Operator may stop providing or change images without notice, such as at the request of rights holders.

Even though "Article 8 (Retrieval from the Website by Automated Means)" permits retrieval from the website, it does not license the use of images on the website.
The conditions in this Article also apply to images retrieved from the website.

## Article 11 (Rate Limits)

Requests to the API and the website are subject to rate limits separately set by the Operator.
The rate limits are listed in "Annict Rate Limits" in the [developer help](https://wikino.app/s/annict/pages/357) (in Japanese).

We may revise the rate limits.
The latest rate limits always apply.

## Article 12 (Transfer of OAuth Applications)

Developers may not transfer a registered OAuth application to a third party.
When handing over an application to another person or legal entity, register a new application in the name of the person taking it over, and have users authorize it again.

This is to prevent users from ending up, without their knowledge, having permitted a different party to read and write their records.

## Article 13 (Display of the Name and Logo)

It is up to you whether your application shows Annict as the source of the data.
We would be happy if you do.
How you show it is also up to you.

You may use the Annict name and logo to show the source of the data or to describe the features of your application (such as "an app for tracking on Annict").

On the other hand, you must not display anything that may lead users to mistake your application for an official application of Annict or one of its partners.
This includes putting the Annict name or logo in the name, icon, or store description of your application.

## Article 14 (Disclaimer of Warranties and Changes to Specifications)

The Operator notifies Developers of the following:

- The specifications of the API and the HTML structure of the website may change without notice
- We may stop providing the API
  When we stop providing a version of the API or the entire API, we will announce it in the developer help or elsewhere at least one year before the end date
- The IDs of works and other items may change or become unavailable due to the merging or deletion of works or the migration between API versions
- The Operator does not guarantee that the information provided by the API is accurate or that the API is always available
- The Operator does not provide support for applications or for retrieval from the website by automated means
  Developers handle any trouble that arises between their applications and their users

The scope of the Operator's liability is governed by the Terms of Service.

## Article 15 (Measures Against Violations)

If a Developer violates these Terms or the Terms of Service, or uses the API in a way that interferes with the operation of the Service, the Operator may take the following measures without prior notice:

- Revoke access tokens
- Delete the registration of OAuth applications
- Block access to the API or the website

## Article 16 (Changes to These Terms)

The Operator may change these Terms through the same procedure as for changing the Terms of Service.
We also change these Terms when the API needs new conditions (such as conditions specific to REST API V2).

## Article 17 (Contact)

Please send requests for individual permission and questions about these Terms to the following email address:

- Email address: hello@annict.com

## Supplementary Provisions

- Established: October 15, 2026
- Effective: November 15, 2026

Purposes for which the Operator individually permitted the use of the API before these Terms took effect will not be newly restricted on the ground that these Terms took effect.
However, this does not include operational conditions such as rate limits.
The latest ones always apply.
