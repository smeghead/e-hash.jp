$(function(event){
  if (typeof(_from) != 'undefined' && _from) {
    console.log('hide h1.');
    $('h1').css('display', 'none');
  }
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
//   var existsInPage = function (element) {
//     var topLocation = 0;
//     do {
//       topLocation += element.offsetTop  || 0;
//       if (element.offsetParent == document.body)
//         if (element.position == 'absolute') break;

//       element = element.offsetParent;
//     } while (element);

//     var scrollPosition = document.documentElement.scrollTop || document.body.scrollTop || 0;

//     var browserHeight = window.innerHeight ||
//       (document.documentElement && document.documentElement.clientHeight && document.documentElement.clientHeight) ||
//       (document.body && document.body.clientHeight) || 0;
//     return (topLocation > scrollPosition) && (topLocation < scrollPosition + browserHeight);
//   };

  //initialize
  try {
    var initialize = function(target){
      $('.create_at', target).each(function(){
        var timestamp = $(this).text();
        timestamp = timestamp.replace(/ \+.*$/, '');
//         var d = new Date();
//         d.setTime(timestamp.substring(0, timestamp.length - 3));
//         var date_string = 
//           zero(d.getFullYear(), 4) + '年' +
//           zero((d.getMonth() + 1), 2) + '月' +
//           zero(d.getDate(), 2) + '日 ' +
//           zero(d.getHours(), 2) + ':' +
//           zero(d.getMinutes(), 2);
        $(this).html(
          '<a target="_blank" href="http://twitter.com/#!/' +
          $(this).data('screenname') + '/status/' +
          $(this).data('statusid') + '">' +
          timestamp + '</a>');
        $(this).show();
      });
      $('a.subject_link', target).each(function(){
        var $this = $(this);
        var hashtag = $this.data('hashtag');
        $this.attr('href', '/s/' + encodeURIComponent(hashtag.substring(1)));
      });
      $('.text', target).each(function(){
        var html = $(this).html();
        var url_regexp = /(ftp|http|https):\/\/(\w+:{0,1}\w*@)?(\S+)(:[0-9]+)?(\/|\/([\w#!:.?+=&%@!\-\/]))?/g;
        html = html.replace(url_regexp, function(x){
          return '<a class="twitter-url" href="' + x + '" target="_blank" data-ajax="false">' + x + '</a>';
        });
        var hashtag_regexp = /[#＃][^ .;:　\n]+/g;
        html = html.replace(hashtag_regexp, function(x){
          return '<a class="twitter-hashtag" href="/s/' + encodeURIComponent(x.substring(1)) + '" data-ajax="false">' + x + '</a>';
        });
        var at_regexp = /@[_a-z0-9]+/ig;
        html = html.replace(at_regexp, function(x){
          return '<a class="twitter-at" href="http://twitter.com/#!/' + encodeURIComponent(x.substring(1)) + '" target="_blank" data-ajax="false">' + x + '</a>';
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
            if (data == 'needs_oauth') {
              document.location.href = '/get_request_token';
              return;
            }
            $('div.retweet', $this)
              .removeClass('retweet')
              .addClass('retweeted');
            // update page.
            if (data != '') {
              var image = document.createElement('img');
              image.src = 'https://api.twitter.com/1/users/profile_image?screen_name=' + data + '&size=mini';
              $('.users', $this.closest('div.subject_block')).append(image);
              message_display('Retweetしました');
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
            if (data == 'needs_oauth') {
              document.location.href = '/get_request_token';
              return;
            }
            $('div.favorite', $this)
              .removeClass('favorite')
              .addClass('favorited');
            // update page.
            if (data != '') {
              var image = document.createElement('img');
              image.src = 'https://api.twitter.com/1/users/profile_image?screen_name=' + data + '&size=mini';
              $('.users', $this.closest('div.subject_block')).append(image);
              message_display('Favoriteしました');
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
  } catch (e) {
    alert('エラーが発生しました。' + e.toString());
  }
  initialize(document);

  $('div.more-tweets').click(function(){
    var $this = $(this);
    var hashtag = $this.data('hashtag');
    var page = $this.data('page') + 1;
    var sort = document.location.search && document.location.search.indexOf('sort=new') > -1 ? 'new' : '';
    $.get('/s/more', {hashtag: hashtag, page: page, sort: sort}, function(data){
      data = data.replace(/^\s+/g, '');
      if (!data ||
          (data.indexOf('<ul') > -1 && data.indexOf('<li') == -1)) { // condition of mobile versoin.
        $('div.more-tweets').hide('slow');
        return;
      }
      var data_element = $(data);
      initialize(data_element);
      if (typeof($.fn.jqmData) != 'undefined') {
        data_element.appendTo( ".ui-page" ).trigger( "create" );
      }

      var blocks = $('div#subject_blocks');
      blocks.append(data_element);
      if ($('ul', blocks).length > 0) {
        $('ul', blocks).listview();
      }
      $this.data('page', page);
      $this.text('もっと読む')
    });
  });
//   $('div.more-tweets').each(function(){
//     var $this = $(this);
//     setInterval(function(){
//       var org_text = $this.text();
//       if (existsInPage($this.get(0))) {
//         if ($this.text() == '読み込み中...')
//           return;
//         $this.text('読み込み中...');
//         $this.click();
//       }
//     }, 1000);
//   });
  $('div.hashtags-more a').click(function(){
    var $this = $(this);
    var page = $this.closest('.hashtags-more').prev('div').data('page') + 1;

    $.get('/hashtags_more', {page: page}, function(data){
      var more = $(data);
      if ($('a', more).length == 0) {
        $this.hide();
        return;
      }
      initialize(more);
      more.insertBefore($this.closest('.hashtags-more'));
//       $('a.subject_link', more).tagcloud();
    });
    return false;
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
//   if ($.fn.tagcloud) {
//     (function(){
//       $.fn.tagcloud.defaults = {
//           size: {start: 14, end: 23, unit: 'pt'}
//           //color: {start: '#cde', end: '#f52'}
//       };
//       $('a.subject_link').tagcloud();
//     })();
//   }
  $('a#contact-to').attr('href', 'mailto:smeghead7+e-hash.jp@gmail.com?subject=' + document.location.hostname + 'についての問合せ');

  //ticker
  (function(){
    if ($('h1 a').length > 0) {
      var ticker_width =
        $('h1').css('width').replace('px', '') -
        $('h1 a').css('width').replace('px', '') - 10;
      $('div#social-bookmarks').css('width', ticker_width + 'px');
      $('div#ticker').css('width', ticker_width + 'px');
      $('div#ticker').show();
    }
  })();
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
