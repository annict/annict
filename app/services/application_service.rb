# frozen_string_literal: true

class ApplicationService
  def initialize
    @result = ServiceResult.new(errors: [])
  end
end
