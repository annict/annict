# frozen_string_literal: true
# == Schema Information
#
# Table name: db_activities
#
#  id                 :integer          not null, primary key
#  user_id            :integer          not null
#  trackable_id       :integer          not null
#  trackable_type     :string           not null
#  action             :string           not null
#  parameters         :json
#  created_at         :datetime         not null
#  updated_at         :datetime         not null
#  root_resource_id   :integer
#  root_resource_type :string
#  object_id          :integer
#  object_type        :string
#
# Indexes
#
#  index_db_activities_on_object_id_and_object_type                (object_id,object_type)
#  index_db_activities_on_root_resource_id_and_root_resource_type  (root_resource_id,root_resource_type)
#  index_db_activities_on_trackable_id_and_trackable_type          (trackable_id,trackable_type)
#

class DbActivity < ActiveRecord::Base
  extend Enumerize

  belongs_to :object, polymorphic: true
  belongs_to :root_resource, polymorphic: true
  belongs_to :trackable, polymorphic: true
  belongs_to :user

  def diffs(new_resource, old_resource)
    HashDiff.diff(old_resource.to_diffable_hash, new_resource.to_diffable_hash)
  end

  def root_resource_action?
    [
      "characters.create",
      "characters.update",
      "organizations.create",
      "organizations.update",
      "people.create",
      "people.update",
      "works.create",
      "works.update"
    ].include?(action)
  end

  def child_resource_action?
    [
      "casts.create",
      "casts.update",
      "episodes.create",
      "episodes.update",
      "programs.create",
      "programs.update",
      "staffs.create",
      "staffs.update"
    ].include?(action)
  end

  def action_table_name
    action&.split(".")&.first
  end

  def action_verb
    action&.split(".")&.last
  end

  def anchor
    return "" if object_type != "DbComment"
    object.anchor
  end
end
