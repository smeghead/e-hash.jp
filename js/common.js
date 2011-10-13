$(function(){
  var zero = function(s, n){
    s = '0000' + s;
    return s.substring(s.length - n);
  };
  var message_display = function(mes){
    var box = $(document.createElement('div'));
    box.addClass('message');
    box.text(mes);
    $('body').append(box);
    $('.message').show('slow');
    setTimeout(function(){
      $('.message').hide('slow');
      $('body').remove('.message');
    }, 5000);
  };
  var existsInPage = function (element) {
    var topLocation = 0;
    do {
      topLocation += element.offsetTop  || 0;
      if (element.offsetParent == document.body)
        if (element.position == 'absolute') break;

      element = element.offsetParent;
    } while (element);

    var scrollPosition = document.documentElement.scrollTop || document.body.scrollTop || 0;

    var browserHeight = window.innerHeight ||
      (document.documentElement && document.documentElement.clientHeight && document.documentElement.clientHeight) ||
      (document.body && document.body.clientHeight) || 0;
    return (topLocation > scrollPosition) && (topLocation < scrollPosition + browserHeight);
  };

  //initialize
  var initialize = function(target){
    $('.create_at', target).each(function(){
      var timestamp = $(this).text();
      var d = new Date();
      d.setTime(timestamp.substring(0, timestamp.length - 3));
      var date_string = 
        zero(d.getFullYear(), 4) + '年' +
        zero((d.getMonth() + 1), 2) + '月' +
        zero(d.getDate(), 2) + '日 ' +
        zero(d.getHours(), 2) + ':' +
        zero(d.getMinutes(), 2);
      $(this).html(
        '<a target="_blank" href="http://twitter.com/#!/' +
        $(this).data('screenname') + '/status/' +
        $(this).data('statusid') + '">' +
        date_string + '</a>');
      $(this).show();
    });
    $('a.subject_link', target).each(function(){
      var $this = $(this);
      var hashtag = $this.data('hashtag');
      $this.attr('href', '/s/' + encodeURIComponent(hashtag.substring(1)));
    });
    $('div.text', target).each(function(){
      var html = $(this).html();
      var url_regexp = /(ftp|http|https):\/\/(\w+:{0,1}\w*@)?(\S+)(:[0-9]+)?(\/|\/([\w#!:.?+=&%@!\-\/]))?/g;
      html = html.replace(url_regexp, function(x){
        return '<a class="twitter-url" href="' + x + '" target="_blank">' + x + '</a>';
      });
      var hashtag_regexp = /[#＃][^ .;:　\n]+/g;
      html = html.replace(hashtag_regexp, function(x){
        return '<a class="twitter-hashtag" href="/s/' + encodeURIComponent(x.substring(1)) + '">' + x + '</a>';
      });
      var at_regexp = /@[_a-z0-9]+/ig;
      html = html.replace(at_regexp, function(x){
        return '<a class="twitter-at" href="http://twitter.com/#!/' + encodeURIComponent(x.substring(1)) + '" target="_blank">' + x + '</a>';
      });
      $(this).html(html);
    });
    //events
    $('a.profile', target).click(function(){
      var $this = $(this);
      if (!type) return false;
      $.post('/point_up', {type: 'profile', key: $this.data('key')},
        function(data){
        }
      );
      return false;
    });
    $('a.retweet', target).click(function(){
      var $this = $(this);
      // increment point.
      $.post('/retweet', {statusId: $this.data('statusid').replace(':', ''), key: $this.data('key')},
        function(data){
          $('div.retweet', $this)
            .removeClass('retweet')
            .addClass('retweeted');
          message_display('Retweetしました');
          // update page.
          if (data != '') {
            var image = document.createElement('img');
            image.src = 'http://img.tweetimag.es/i/' + data + '_m';
            $('.users', $this.parent().parent().parent()).append(image);
          }
        }
      );
      return false;
    });
    $('a.favorite', target).click(function(){
      var $this = $(this);
      // increment point.
      $.post('/favorite', {statusId: $this.data('statusid').replace(':', ''), key: $this.data('key')},
        function(data){
          $('div.favorite', $this)
            .removeClass('favorite')
            .addClass('favorited');
          message_display('Favoriteしました');
          // update page.
          if (data != '') {
            var image = document.createElement('img');
            image.src = 'http://img.tweetimag.es/i/' + data + '_m';
            $('.users', $this.parent().parent().parent()).append(image);
          }
        }
      );
      return false;
    });
    $('div.like a.like', target).click(function(){
      var $this = $(this);
      $.post('/like', {key: $this.data('key'), url: document.location.pathname},
        function(data){
          if (data == 'needs_oauth') {
            document.location.href = '/get_request_token';
            return;
          }
          // update page.
          if (data != '') {
            var image = document.createElement('img');
            image.src = 'http://img.tweetimag.es/i/' + data + '_m';
            $('.users', $this.parent().parent().parent().parent()).append(image);
            message_display('Like!しました');
          }
        }
      );
      return false;
    });
    // user count.
    $('div.subject_block', target).each(function(){
      var $this = $(this);
      var count = $('div.users img', $this).length;
      if (count > 0) {
        $('span.user-count', $this).text(count + ' ');
      }
    });
  };
  initialize(document);

  $('div.more-tweets').click(function(){
    var $this = $(this);
    var hashtag = $this.data('hashtag');
    var page = $this.data('page') + 1;
    var sort = document.location.search && document.location.search.indexOf('sort=new') > -1 ? 'new' : '';
    $.get('/s/more', {hashtag: hashtag, page: page, sort: sort}, function(data){
      if (!data) {
        $('div.more-tweets').hide('slow');
        return;
      }
      var data_element = $(data);
      initialize(data_element);
      $('div#subject_blocks').append(data_element);
      $this.data('page', page);
      $this.text('もっと読む')
    });
  });
  $('div.more-tweets').each(function(){
    var $this = $(this);
    setInterval(function(){
      var org_text = $this.text();
      if (existsInPage($this.get(0))) {
        $this.text('読み込み中...');
        $this.click();
      }
    }, 1000);
  });

  //ticker
  var ticker_width =
    $('h1').css('width').replace('px', '') -
    $('h1 a').css('width').replace('px', '') - 10;
  $('div#social-bookmarks').css('width', ticker_width + 'px');
  $('div#ticker').css('width', ticker_width + 'px');
  $('div#ticker-elements').html($('div#ticker-elements-hide').html());
  $('div#ticker-elements').jStockTicker({speed: 2, interval: 40});
  $('div#ticker-elements a').click(function(){
    var hashtag = $(this).text();
    document.location.href = '/s/' + hashtag.substring(1);
  });

  //tweet
  var additional_text = ' ' + $('#hashtag').text() + ' http://www.e-hash.jp/';
  var button = $('div#tweet-box input.post');
  var message = $('#leave-letter-count');
  message.text(140 - additional_text.length);
  $('div#tweet-box textarea').keyup(function(){
    var status_len = $(this).val().length;
    var len = 140 - additional_text.length - status_len;
    message.text(len);
    if (status_len != 0 && len >= 0) {
      button.removeAttr('disabled');
    } else {
      button.attr('disabled', 'disabled');
    }
  });
  $('div#tweet-box input.post').click(function(){
    var $this = $(this);
    var hashtag = $('#hashtag').text();
    $.post('/post', {
        hashtag: hashtag,
        url: document.location.pathname,
        status: $('#status').val() + ' ' + hashtag + ' http://www.e-hash.jp/'
      },
      function(data){
        if (data == 'needs_oauth') {
          document.location.href = '/get_request_token';
          return;
        }
        // update page.
        if (data != '') {
          // show message
          $this.text('');
          var tweet = $(data);
          initialize(tweet);
          tweet.insertBefore($('div#subject_blocks').first());
          message_display('Twitterにつぶやきました');
        }
      }
    );
    return false;
  });

  // hashtags tagcloud
  if ($.fn.tagcloud) {
    $.fn.tagcloud.defaults = {
        size: {start: 14, end: 23, unit: 'pt'}
        //color: {start: '#cde', end: '#f52'}
    };
    $('a.subject_link').tagcloud();
  }
  $('a#contact-to').attr('href', 'mailto:smeghead7+e-hash.jp@gmail.com?subject=' + document.location.hostname + 'についての問合せ');
});

var _gaq = _gaq || [];
_gaq.push(['_setAccount', 'UA-25384750-1']);
_gaq.push(['_trackPageview']);

(function() {
  var ga = document.createElement('script'); ga.type = 'text/javascript'; ga.async = true;
  ga.src = ('https:' == document.location.protocol ? 'https://ssl' : 'http://www') + '.google-analytics.com/ga.js';
  var s = document.getElementsByTagName('script')[0]; s.parentNode.insertBefore(ga, s);
})();

//  vim: set ts=2 sw=2 sts=2 expandtab fenc=utf-8:
