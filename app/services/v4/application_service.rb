# frozen_string_literal: true

module V4
  class ApplicationService
    def initialize
      @result = result_class.new(errors: [])
    end

    private

    def result_class
      raise NotImplementedError
    end
  end
end
