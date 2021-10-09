# frozen_string_literal: true

module Updaters
  class NoteUpdater
    attr_reader :library_entry

    def initialize(form:)
      @form = form
      @library_entry = @form.library_entry
    end

    def call
      @library_entry.update!(note: @form.body)

      self
    end
  end
end
