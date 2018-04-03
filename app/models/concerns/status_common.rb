# frozen_string_literal: true

module StatusCommon
  extend ActiveSupport::Concern

  included do
    extend Enumerize

    enumerize :kind, scope: true, in: {
      wanna_watch: 1,
      watching: 2,
      watched: 3,
      on_hold: 5,
      stop_watching: 4
    }

    belongs_to :user
    belongs_to :work

    scope :positive, -> { with_kind(:wanna_watch, :watching, :watched) }
    scope :work_published, -> { joins(:work).merge(Work.published) }
  end
end
