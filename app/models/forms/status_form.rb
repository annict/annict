# frozen_string_literal: true

module Forms
  class StatusForm < Forms::ApplicationForm
    attr_accessor :anime
    attr_reader :kind

    validates :kind, inclusion: { in: Status::KIND_MAPPING.values.map(&:to_s) }

    def kind=(value)
      @kind = value&.to_s&.downcase.presence
    end

    def no_status?
      @kind == 'no_status'
    end
  end
end
