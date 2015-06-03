module WorkCommon
  extend ActiveSupport::Concern

  included do
    extend Enumerize

    enumerize :media, in: { tv: 1, ova: 2, movie: 3, web: 4, other: 0 }

    belongs_to :season

    validates :sc_tid, numericality: { only_integer: true }, allow_blank: true,
                       uniqueness: true
    validates :title, presence: true, uniqueness: true
    validates :media, presence: true
    validates :official_site_url, url: { allow_blank: true }
    validates :wikipedia_url, url: { allow_blank: true }
  end
end
