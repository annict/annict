# == Schema Information
#
# Table name: collection_item_agreements
#
#  id                 :integer          not null, primary key
#  user_id            :integer          not null
#  collection_item_id :integer          not null
#  created_at         :datetime         not null
#  updated_at         :datetime         not null
#
# Indexes
#
#  index_cia_on_uid_and_ciid                               (user_id,collection_item_id) UNIQUE
#  index_collection_item_agreements_on_collection_item_id  (collection_item_id)
#  index_collection_item_agreements_on_user_id             (user_id)
#

class CollectionItemAgreement < ApplicationRecord
end
