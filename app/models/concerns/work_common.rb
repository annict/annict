module WorkCommon
  extend ActiveSupport::Concern

  DIFF_FIELDS = %i(season_id sc_tid title media official_site_url wikipedia_url
                   released_at twitter_username twitter_hashtag released_at_about)
  PUBLISH_FIELDS = DIFF_FIELDS

  included do
    extend Enumerize

    enumerize :media, in: { tv: 1, ova: 2, movie: 3, web: 4, other: 0 }

    belongs_to :season

    validates :media, presence: true
    validates :official_site_url, url: { allow_blank: true }
    validates :wikipedia_url, url: { allow_blank: true }

    def to_diffable_hash
      data = self.class::DIFF_FIELDS.inject({}) do |hash, field|
        hash[field] = case field
        when :media
          send(field).to_s
        when :released_at
          send(field).try(:strftime, "%Y/%m/%d")
        else
          send(field)
        end

        hash
      end

      data.delete_if { |_, v| v.blank? }
    end
  end
end
