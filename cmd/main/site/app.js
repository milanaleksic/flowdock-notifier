var app = new Vue({
    el: '#app',
    data: $.extend(igor.config, {
        email: "",
        id: "",
        name: ""
    }),
    computed: {
        welcomeMessage: function () {
            if (this.email) {
                return "Hi " + this.name
            } else {
                return "Checking..."
            }
        }
    },
    methods: {
        // called by Google API when signing in has been done
        googleSignIn: function (googleUser) {
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
            return gapi.auth2.getAuthInstance().signIn({
                prompt: 'login'
            }).then(function (newUser) {
                app.updateCognitoServerToUseGoogleUser(newUser);
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
        }
    }
})

googleSignIn = app.googleSignIn;