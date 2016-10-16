$.extend(igor, {
    identity: new $.Deferred(),

    googleSignIn: function (googleUser) {
        AWS.config.update({
            region: igor.config.awsRegion,
            credentials: new AWS.CognitoIdentityCredentials({
                IdentityPoolId: igor.config.cognitoPoolId,
                Logins: {
                    'accounts.google.com': googleUser.getAuthResponse().id_token
                }
            })
        });
        function refresh() {
            return gapi.auth2.getAuthInstance().signIn({
                prompt: 'login'
            }).then(function (userUpdate) {
                var creds = AWS.config.credentials;
                var newToken = userUpdate.getAuthResponse().id_token;
                creds.params.Logins['accounts.google.com'] = newToken;
                return igor.awsRefresh();
            });
        }
        igor.awsRefresh().then(function (id) {
            igor.identity.resolve({
                id: id,
                email: googleUser.getBasicProfile().getEmail(),
                refresh: refresh
            });
        });
    },

    awsRefresh: function () {
        var deferred = new $.Deferred();
        AWS.config.credentials.refresh(function (err) {
            if (err) {
                deferred.reject(err);
            } else {
                deferred.resolve(AWS.config.credentials.identityId);
            }
        });
        return deferred.promise();
    },

    appOnReady: function () {
        igor.identity.done(function (profile) {
            console.log("You are logged in as ", profile.email)
        });
    }
});

$(window).ready(igor.appOnReady);

googleSignIn= function (googleUser) {
    igor.googleSignIn(googleUser)
}