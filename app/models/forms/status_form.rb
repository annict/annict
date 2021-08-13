# frozen_string_literal: true

module Forms
  class StatusForm < Forms::ApplicationForm
    attr_accessor :work
    attr_reader :kind

    validates :kind, inclusion: {in: Status::KIND_MAPPING.values.map(&:to_s)}

    def kind=(value)
      @kind = value&.to_s&.downcase.presence

      if @kind
        # v2からv3に変換する
        @kind = Status::KIND_MAPPING[@kind.to_sym]&.to_s.presence || @kind
      end
    end

    def no_status?
      @kind == "no_status"
    end
  end
end
