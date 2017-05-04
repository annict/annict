# == Schema Information
#
# Table name: collection_items
#
#  id            :integer          not null, primary key
#  collection_id :integer          not null
#  work_id       :integer          not null
#  title         :string           not null
#  comment       :text
#  created_at    :datetime         not null
#  updated_at    :datetime         not null
#
# Indexes
#
#  index_collection_items_on_collection_id              (collection_id)
#  index_collection_items_on_collection_id_and_work_id  (collection_id,work_id) UNIQUE
#  index_collection_items_on_work_id                    (work_id)
#

class CollectionItem < ApplicationRecord
end
