# frozen_string_literal: true

module Forms
  class RecordTextareaComponent < ApplicationComponent
    def initialize(form:, textarea_name:, optional_textarea_classname: "")
      @form = form
      @textarea_name = textarea_name
      @optional_textarea_classname = optional_textarea_classname
    end

    def textarea_classname
      @textarea_classname = %w()
      @textarea_classname << @optional_textarea_classname
      @textarea_classname.join(" ")
    end
  end
end
