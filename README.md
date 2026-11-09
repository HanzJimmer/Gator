gator is an RSS feed compiler that will allow you to
    - register users
    - have individual users add feeds to their profiles
    - have users follow other users' feeds
    - regularly update feeds
    - browse posts within feeds
    - and more!

To run this program, you will need both Postgres and Go (1.26.5) installed onto your computer 

after running go install /gator.exe, you will need to set up the config file and database schema. Once complete, the following commands are available
    *reset - resets the full table to a state with no users/feeds
    *register [username] - creates a user
    *login [username] - sets the current user
    *users - lists all users
    *addfeed [url] - creates a feed owned by current user located at the specified url
    *feeds - lists all feeds, its url, and the owner
    *follow [url] - follows a feed for current user at a given url
    *unfollow [url] - unfollows a feed for current user at the given url
    *following - lists all feeds current user is following
    *agg [refreshrate] - begins a feed refesher that will update each tick of your refresh rate
    *browse [optional number]- lists posts within a feed up to given number. defaults to 2