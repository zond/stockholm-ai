

SVG = "http://www.w3.org/2000/svg";

function selEscape(sel) {
  return sel.replace(/\//g, "\\/").replace(/=/g, '\\=');
}

function uniqueColors(numColors) {
  if (numColors == 1) {
	  return ["#ff0000"];
	}
	var result = [];
	var sat = 0.8;
	for (var i = 0; i < numColors; i++) {
	  var hue = 360.0 * i / numColors;
		if (i % 2 == 0) {
		  result.push(tinycolor({
			  h: hue,
				s: sat,
				v: 0.9,
			}).toHexString());			
		} else {
		  result.push(tinycolor({
			  h: hue,
				s: sat,
				v: 0.7,
			}).toHexString());
		}
	}
	return result;
}

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
			"games": "games",
			"games/:id": "showGame",
		},

		showGame: function(id) {
		  this.render(new GameView({ id: id }));
		},

		games: function() {
		  this.render(new GamesView({}));
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

