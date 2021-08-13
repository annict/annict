# frozen_string_literal: true

module Annict
  module DataCare
    class UpdateNextResourcesOnLibraryEntries
      def self.run!
        new.run!
      end

      def run!
        target_library_entries.find_each do |le|
          puts "library_entries.id: #{le.id}"

          le.set_next_resources!
          le.save!
        end
      end

      private

      def target_library_entries
        @target_library_entries ||= LibraryEntry
          .preload(:work, :program)
          .watching
          .after(Date.today - 3, field: "library_entries.updated_at")
          .merge(LibraryEntry.has_no_next_episode.or(LibraryEntry.has_no_next_slot))
      end
    end
  end
end
