# typed: false
# frozen_string_literal: true

class DbActivity < ApplicationRecord
  extend Enumerize

  belongs_to :object, polymorphic: true, optional: true
  belongs_to :root_resource, polymorphic: true, optional: true
  belongs_to :trackable, polymorphic: true
  belongs_to :user

  def diffs(new_resource, old_resource)
    Hashdiff.diff(old_resource.to_diffable_hash, new_resource.to_diffable_hash)
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
      "slots.create",
      "slots.update",
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
