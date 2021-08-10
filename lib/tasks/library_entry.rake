# frozen_string_literal: true

namespace :library_entry do
  task update_next_resources: :environment do
    library_entries = LibraryEntry
      .preload(:anime, :program)
      .watching
      .after(Date.today - 3, field: "library_entries.updated_at")
      .merge(LibraryEntry.has_no_next_episode.or(LibraryEntry.has_no_next_slot))

    library_entries.find_each do |le|
      puts "library_entries.id: #{le.id}"

      le.set_next_resources!
      le.save!
    end
  end
end
