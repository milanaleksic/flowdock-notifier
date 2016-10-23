var app = new Vue({
    el: '#app',
    data: $.extend(igor.config, {
        // From Google
        email: "",
        id: "",
        name: "",
        // UX
        savedFlash: false,
        eligibleForConfiguration: false,
        notSignedIn: false,
        // From Dynamo
        message: "",
        activeFrom: "",
        activeUntil: "",
        flowdockUsername: "",
        flowdockToken: "",
    }),
    mounted: function () {
        var that = this;
        gapi.load('auth2', function () {
            auth2 = gapi.auth2.init({
                client_id: igor.config.googleClientId
            });
            auth2.then(function () {
                if (!auth2.isSignedIn.get()) {
                    that.notSignedIn = true;
                } else {
                    that.googleSignedIn(auth2.currentUser.get());
                }
            });
        });
    },
    computed: {
        welcomeMessage: function () {
            var that = this;
            if (this.email) {
                var db = new AWS.DynamoDB.DocumentClient();
                db.get({
                    TableName: 'igor',
                    Key: {
                        userId: this.id,
                    }
                }, function (err, data) {
                    if (err) {
                        console.error("Unable to get the data. Error JSON:", JSON.stringify(err, null, 2));
                    } else {
                        that.notSignedIn = false;
                        that.eligibleForConfiguration = true;
                        that.flowdockToken = data.Item.flowdockToken;
                        that.flowdockUsername = data.Item.flowdockUsername;
                        that.message = data.Item.message;
                        that.activeFrom = data.Item.activeFrom;
                        that.activeUntil = data.Item.activeUntil;
                    }
                });
                return "Hi " + this.name
            } else {
                return ""
            }
        }
    },
    methods: {
        // called when signing in has been done or verification we are signed has been done
        googleSignedIn: function (googleUser) {
            AWS.config.update({
                region: this.awsRegion,
                credentials: new AWS.CognitoIdentityCredentials({
                    IdentityPoolId: this.cognitoPoolId,
                    Logins: {
                    }
                })
            });
            this.updateCognitoServerToUseGoogleUser(googleUser);
        },
        // this is called by us when we detect credentials are not OK any more
        refreshGoogleCredentials: function () {
            var that = this;
            return gapi.auth2.getAuthInstance().signIn({
                prompt: 'login'
            }).then(function (newUser) {
                that.googleSignedIn(newUser);
            })
        },
        updateCognitoServerToUseGoogleUser: function (googleUser) {
            var appThis = this
            AWS.config.credentials.params.Logins['accounts.google.com'] = googleUser.getAuthResponse().id_token;
            AWS.config.credentials.refresh(function (err) {
                if (err) {
                    console.error("Error: ", err);
                } else {
                    appThis.email = googleUser.getBasicProfile().getEmail();
                    appThis.name = googleUser.getBasicProfile().getName();
                    appThis.id = AWS.config.credentials.identityId;
                }
            })
        },
        // called from HTML when "save" is clicked
        saveConfiguration: function () {
            if (this.email) {
                var db = new AWS.DynamoDB.DocumentClient();
                var that = this;
                db.put({
                    TableName: 'igor',
                    Item: {
                        userId: this.id,
                        eligibleForConfiguration: this.eligibleForConfiguration,
                        flowdockToken: this.flowdockToken,
                        flowdockUsername: this.flowdockUsername,
                        message: this.message,
                        activeFrom: this.activeFrom,
                        activeUntil: this.activeUntil
                    }
                }, function (err, data) {
                    if (err) {
                        console.error("Unable to save the data. Error JSON:", JSON.stringify(err, null, 2));
                    } else {
                        that.savedFlash = true;
                        window.setTimeout(function () { that.savedFlash = false }, 2000)
                    }
                });
                return "Hi " + this.name
            } else {
                return "Checking..."
            }
        },
        // called from HTML when "sign out" is clicked
        signOut: function () {
            var that = this;
            gapi.auth2.getAuthInstance().signOut().then(function () {
                that.email = ""
                that.id = "";
                that.name = "";
                that.eligibleForConfiguration = false;
                that.notSignedIn = true;
            });
        }
    }
})
