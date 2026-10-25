# typed: false
# frozen_string_literal: true

class DbActivity < ApplicationRecord
  extend Enumerize

  belongs_to :object, polymorphic: true, optional: true
  belongs_to :root_resource, polymorphic: true, optional: true
  belongs_to :trackable, polymorphic: true
  belongs_to :user

  # 人物が非公開・削除済みの編集履歴を除く。差分に人物名が出るため、その人物を参照するキャスト・スタッフの履歴も除く
  scope :with_kept_person, -> {
    hidden_person_ids = Person.where.not(id: Person.only_kept.select(:id)).select(:id)

    where.not(trackable_type: "Person", trackable_id: hidden_person_ids)
      .where.not(trackable_type: "Cast", trackable_id: Cast.where(person_id: hidden_person_ids).select(:id))
      .where.not(trackable_type: "Staff", trackable_id: Staff.where(resource_type: "Person", resource_id: hidden_person_ids).select(:id))
  }

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
