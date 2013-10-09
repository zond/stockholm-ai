window.session = {};

$(window).load(function() {
  
	$(document).on('click', 'a.navigate', function(ev) {
	  ev.preventDefault();
		window.session.nav.navigate($(ev.currentTarget).attr('href'));
	});

	var AppRouter = Backbone.Router.extend({
	
		currentView: null,

		render: function(view) {
			if (this.currentView != null) {
			  this.currentView.remove();
			}
			$('#content').append(view.render().el);
			this.currentView = view;
		},

		routes: {
			"": "about",
			"ais": "ais",
		},

		ais: function() {
			this.render(new AIsView({}));
		},

		about: function() {
			this.render(new AboutView({}));
		},
	});	
	
	window.session.user = new User();
	window.session.user.fetch();

	window.session.nav = new TopNavigationView({
		el: $('nav'),
	}).render();

	window.session.router = new AppRouter();
	Backbone.history.start({ 
		pushState: true,
	});

	window.session.nav.navigate(Backbone.history.fragment || '/');
});

