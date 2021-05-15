# frozen_string_literal: true

module V4
  class EmptyComponent < V4::ApplicationComponent
    def initialize(text:)
      @text = text
    end

    private

    attr_reader :text
  end
end
