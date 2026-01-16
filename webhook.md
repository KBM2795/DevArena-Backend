Sync Clerk data to your app with webhooks
Before you start
A Clerk app is required.
A ngrok account is required.
In some cases, you may want to sync Clerk's user table to a user table in your own database. Read the next few sections carefully to determine if this is the right approach for your app.

When to sync Clerk data
Syncing data with webhooks can be a suitable approach for some applications, but it comes with important considerations. Webhook deliveries are not guaranteed and may occasionally fail due to problems like network issues, so your implementation should be prepared to handle retries and error scenarios. Additionally, syncing data via webhooks is eventually consistent, meaning there can be a delay between when a Clerk event (such as a user being created or updated) occurs and when the corresponding data is reflected in your database. If not managed carefully, this delay can introduce bugs and race conditions.

If you can access the necessary data directly from the Clerk session token, you can achieve strong consistency while avoiding the overhead of maintaining a separate user table in your own database and the latency of retrieving that data on every request. This makes not syncing data much more efficient, if your use case allows for it.

The most notable use case for syncing Clerk data is if your app has social features where users can see content posted by other users. This is because Clerk's frontend API only allows you to access information about the currently signed-in user. If your app needs to display information about other users, like their names or avatars, you can't access that data from the frontend API alone. While you can fetch other users' data using Clerk's backend API for each request, this is slow compared to a database lookup, and you risk hitting rate limits. In this case, it makes sense to store user data in your own database and sync it from Clerk.

Storing extra user data
If you want to use webhooks to sync Clerk data because you want to store extra data for the user, consider the following approaches:

(Recommended) If it's more than 1.2KB, you could store only the extra user data in your own database.

Store the user's Clerk ID as a column in the users table in your own database, and only store extra user data. When you need to access Clerk user data, access it directly from the Clerk session token. When you need to access the extra user data, do a lookup in your database using the Clerk user ID. Consider indexing the Clerk user ID column since it will be used frequently for lookups.
For example, Clerk doesn't collect a user's birthday, country, or bio, but if you wanted to collect these fields, you could store them in your own database like this:
id	clerk_id	birthday	country	bio
user123abc	user_123	1990-05-12	USA	Coffee enthusiast.
user456abc	user_456	1985-11-23	Canada	Loves to read.
user789abc	user_789	2001-07-04	Germany	Student and coder.
If it's less than 1.2KB, you could use Clerk metadata and store it in the user's session token.

For minimal custom data (under 1.2KB), you can store it in a user's metadata instead of dealing with a separate users table. Then, you can store the metadata in the user's session token to avoid making a network request to Clerk's Backend API when retrieving it. However, if there's any chance that a user will ever have more than 1.2KB of extra data, you should use the other approach, as you risk cookie size overflows if metadata is over 1.2KB.
Another limitation to consider is that metadata cannot be queried, so you can't use it to filter users by metadata. For example, if you stored a user's birthday in metadata, you couldn't find all users with a certain birthday. If you need to query the data that you're storing, you should store it in your own database instead.
A hybrid approach of the two approaches above.

How to sync Clerk data
In this guide, you'll set up a webhook in your app to listen for the user.created event, create an endpoint in the Clerk Dashboard, build a handler for verifying the webhook, and test it locally using ngrok and the Clerk Dashboard.

Clerk offers many events, but three key events include:

user.created: Triggers when a new user registers in the app or is created via the Clerk Dashboard or Backend API. Listening to this event allows the initial insertion of user information in your database.
user.updated: Triggers when user information is updated via Clerk components, the Clerk Dashboard, or Backend API. Listening to this event keeps data synced between Clerk and your external database. It is recommended to only sync what you need to simplify this process.
user.deleted: Triggers when a user deletes their account, or their account is removed via the Clerk Dashboard or Backend API. Listening to this event allows you to delete the user from your database or add a deleted: true flag.
These steps apply to any Clerk event. To make the setup process easier, it's recommended to keep two browser tabs open: one for your Clerk Webhooks page and one for your ngrok dashboard.

Set up ngrok
To test a webhook locally, you need to expose your local server to the internet. This guide uses ngrok which creates a forwarding URL that sends the webhook payload to your local server.

Navigate to the ngrok dashboard to create an account.
On the ngrok dashboard homepage, follow the setup guide instructions. Under Deploy your app online, select Static domain. Run the provided command, replacing the port number with your server's port. For example, if your development server runs on port 3000, the command should resemble ngrok http --url=<YOUR_FORWARDING_URL> 3000. This creates a free static domain and starts a tunnel.
Save your Forwarding URL somewhere secure.
Set up a webhook endpoint
In the Clerk Dashboard, navigate to the Webhooks page.
Select Add Endpoint.
In the Endpoint URL field, paste the ngrok Forwarding URL you saved earlier, followed by /api/webhooks. This is the endpoint that Clerk uses to send the webhook payload. The full URL should resemble https://fawn-two-nominally.ngrok-free.app/api/webhooks.
In the Subscribe to events section, scroll down and select user.created.
Select Create. You'll be redirected to your endpoint's settings page. Keep this page open.
Add your Signing Secret to .env
To verify the webhook payload, you'll need your endpoint's Signing Secret. Since you don't want this secret exposed in your codebase, store it as an environment variable in your .env file during local development.

On the endpoint's settings page in the Clerk Dashboard, copy the Signing Secret. You may need to select the eye icon to reveal the secret.
In your project's root directory, open or create an .env file, which should already include your Clerk API keys. Assign your Signing Secret to CLERK_WEBHOOK_SIGNING_SECRET. The file should resemble:
.env

DevArena

DevArena


NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY=pk_test_cHJvdWQtc3F1aXJyZWwtNC5jbGVyay5hY2NvdW50cy5kZXYk
CLERK_SECRET_KEY=sk_test_G4YhNwfAnlRJ02b74D9szMTY3vZVf5WWBziYZSWaKZ
CLERK_WEBHOOK_SIGNING_SECRET=whsec_123
Make sure the webhook route is public
Incoming webhook events don't contain auth information. They come from an external source and aren't signed in or out, so the route must be public to allow access. If you're using clerkMiddleware(), ensure that the /api/webhooks(.*) route is set as public. For information on configuring routes, see the 
clerkMiddleware() guide
.

Create a route handler to verify the webhook
Set up a Route Handler that uses Clerk's 
verifyWebhook()
 function to verify the incoming Clerk webhook and process the payload.

For this guide, the payload will be logged to the console. In a real app, you'd use the payload to trigger an action. For example, if listening for the user.created event, you might perform a database create or upsert to add the user's Clerk data to your database's user table.

If the route handler returns a 4xx or 5xx code, or no code at all, the webhook event will be retried. If the route handler returns a 2xx code, the event will be marked as successful, and retries will stop.

Note

The following Route Handler can be used for any webhook event you choose to listen to. It is not specific to user.created.

Next.js
Astro
Express
Fastify
Nuxt
React Router
TanStack React Start
app/api/webhooks/route.ts

import { verifyWebhook } from '@clerk/nextjs/webhooks'
import { NextRequest } from 'next/server'

export async function POST(req: NextRequest) {
  try {
    const evt = await verifyWebhook(req)

    // Do something with payload
    // For this guide, log payload to console
    const { id } = evt.data
    const eventType = evt.type
    console.log(`Received webhook with ID ${id} and event type of ${eventType}`)
    console.log('Webhook payload:', evt.data)

    return new Response('Webhook received', { status: 200 })
  } catch (err) {
    console.error('Error verifying webhook:', err)
    return new Response('Error verifying webhook', { status: 400 })
  }
}
Narrow to a webhook event for type inference
WebhookEvent encompasses all possible webhook types. Narrow down the event type for accurate typing for specific events.

In the following example, the if statement narrows the type to user.created, enabling type-safe access to evt.data with autocompletion.

app/api/webhooks/route.ts

console.log(`Received webhook with ID ${id} and event type of ${eventType}`)
console.log('Webhook payload:', body)

if (evt.type === 'user.created') {
  console.log('userId:', evt.data.id)
}
To handle types manually, import the following types from your backend SDK (e.g., @clerk/nextjs/webhooks):

DeletedObjectJSON
EmailJSON
OrganizationInvitationJSON
OrganizationJSON
OrganizationMembershipJSON
SessionJSON
SMSMessageJSON
UserJSON
Test the webhook
Start your Next.js server.
In your endpoint's settings page in the Clerk Dashboard, select the Testing tab.
In the Select event dropdown, select user.created.
Select Send Example.
In the Message Attempts section, confirm that the event's Status is labeled with Succeeded. In your server's terminal where your app is running, you should see the webhook's payload.
Handling failed messages
In the Message Attempts section, select the event whose Status is labeled with Failed.
Scroll down to the Webhook Attempts section.
Toggle the arrow next to the Status column.
Review the error. Solutions vary by error type. For more information, refer to the guide on debugging your webhooks.
Trigger the webhook
To trigger the user.created event, create a new user in your app.

In the terminal where your app is running, you should see the webhook's payload logged. You can also check the Clerk Dashboard to see the webhook attempt, the same way you did when testing the webhook.

Configure your production instance
When you're ready to deploy your app to production, follow the guide on deploying your Clerk app to production.
Create your production webhook by following the steps in the previous Set up a webhook endpoint section. In the Endpoint URL field, instead of pasting the ngrok URL, paste your production app URL.
After you've set up your webhook endpoint, you'll be redirected to your endpoint's settings page. Copy the Signing Secret.
On your hosting platform, update your environment variables on your hosting platform by adding Signing Secret with the key of CLERK_WEBHOOK_SIGNING_SECRET.
Redeploy your app.