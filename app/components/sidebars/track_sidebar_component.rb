# frozen_string_literal: true

module Sidebars
  class TrackSidebarComponent < ApplicationComponent
    def initialize(library_entry_entities:)
      @library_entry_entities = library_entry_entities
    end
  end
end
