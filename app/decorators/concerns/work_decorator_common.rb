module WorkDecoratorCommon
  extend ActiveSupport::Concern

  included do
    def to_values
      model.class::DIFF_FIELDS.inject({}) do |hash, field|
        hash[field] = case field
        when :season_id
          if send(:season_id).present?
            Season.find(send(:season_id)).decorate.yearly_season_ja
          end
        when :sc_tid
          sc_tid = send(:sc_tid)
          if sc_tid.present?
            url = "http://cal.syoboi.jp/tid/#{sc_tid}"
            h.link_to(sc_tid, url, target: "_blank")
          end
        when :media
          Work.media.find_value(send(:media)).text
        when :official_site_url
          url = send(field)
          if url.present?
            h.link_to(url, url, target: "_blank")
          end
        when :wikipedia_url
          url = send(field)
          if url.present?
            h.link_to(URI.decode(url), url, target: "_blank")
          end
        when :twitter_username
          username = send(:twitter_username)
          if username.present?
            url = "https://twitter.com/#{username}"
            h.link_to("@#{username}", url, target: "_blank")
          end
        when :twitter_hashtag
          hashtag = send(:twitter_hashtag)
          if hashtag.present?
            url = "https://twitter.com/search?q=%23#{hashtag}"
            h.link_to("##{hashtag}", url, target: "_blank")
          end
        else
          send(field)
        end

        hash
      end
    end
  end
end
