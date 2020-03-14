# frozen_string_literal: true

class ApplicationComponent < ViewComponent::Base
  class_attribute :bypass_compilation

  def self.inline!
    self.bypass_compilation = true
  end

  def self.compile(raise_template_errors: false)
    super(raise_template_errors: raise_template_errors) unless bypass_compilation
  end
end
