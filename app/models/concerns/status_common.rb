# frozen_string_literal: true

module StatusCommon
  extend ActiveSupport::Concern

  included do
    extend Enumerize

    KIND_MAPPING = {
      wanna_watch: :plan_to_watch,
      watching: :watching,
      watched: :completed,
      on_hold: :on_hold,
      stop_watching: :dropped,
      no_select: :no_status
    }.freeze

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
    scope :work_published, -> { joins(:work).merge(Work.without_deleted) }

    def self.kind_v2_to_v3(kind_v2)
      return if kind_v2.blank?

      KIND_MAPPING[kind_v2.to_sym]
    end

    def self.kind_v3_to_v2(kind_v3)
      return if kind_v3.blank?

      KIND_MAPPING.invert[kind_v3.to_sym]
    end
  end
end
