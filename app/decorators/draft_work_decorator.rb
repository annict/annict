class DraftWorkDecorator < ApplicationDecorator
  def to_values
    model.class::DIFF_FIELDS.inject({}) do |hash, field|
      hash[field] = case field
      when :season_id
        Season.find(send(:season_id)).name
      when :sc_tid
        sc_tid = send(:sc_tid)
        url = "http://cal.syoboi.jp/tid/#{sc_tid}"
        h.link_to(sc_tid, url, target: "_blank")
      when :media
        Work.media.find_value(send(:media)).text
      when :official_site_url
        h.link_to(send(field), send(field), target: "_blank")
      when :wikipedia_url
        h.link_to(URI.decode(send(field)), send(field), target: "_blank")
      when :released_at
        send(:released_at).try(:strftime, "%Y/%m/%d")
      when :twitter_username
        username = send(:twitter_username)
        url = "https://twitter.com/#{username}"
        h.link_to("@#{username}", url, target: "_blank")
      when :twitter_hashtag
        hashtag = send(:twitter_hashtag)
        url = "https://twitter.com/search?q=%23#{hashtag}"
        h.link_to("##{hashtag}", url, target: "_blank")
      else
        send(field)
      end

      hash
    end
  end
end
