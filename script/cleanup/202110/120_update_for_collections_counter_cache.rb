# frozen_string_literal: true

Collection.find_each do |collection|
  puts "Collection: #{collection.id}"
  collection.update_columns(collection_items_count: collection.collection_items.count)
end
