module WorkCommon
  extend ActiveSupport::Concern

  included do
    extend Enumerize

    enumerize :media, in: { tv: 1, ova: 2, movie: 3, web: 4, other: 0 }

    DIFF_FIELDS = %i(season_id sc_tid title media official_site_url wikipedia_url
                     released_at twitter_username twitter_hashtag released_at_about)

    belongs_to :season

    validates :sc_tid, numericality: { only_integer: true }, allow_blank: true,
                       uniqueness: true
    validates :title, presence: true, uniqueness: true
    validates :media, presence: true
    validates :official_site_url, url: { allow_blank: true }
    validates :wikipedia_url, url: { allow_blank: true }

    def to_diffable_hash
      self.class::DIFF_FIELDS.inject({}) do |hash, field|
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
    end
  end
end
