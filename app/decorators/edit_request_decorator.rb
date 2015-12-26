class EditRequestDecorator < ApplicationDecorator
  def status_label
    if opened?
      h.content_tag :span, "オープン", class: "label label-success c-label--status"
    elsif published?
      h.content_tag :span, "公開済み", class: "label label-info c-label--status"
    elsif closed?
      h.content_tag :span, "クローズ", class: "label label-danger c-label--status"
    end
  end

  def resource_label
    text = case kind
    when :draft_work
      "作品"
    when :draft_cast, :draft_person
      "スタッフ / キャスト"
    end

    h.content_tag :span, text, class: "label label-default c-label--transparent"
  end

  def resource_diff_table
    data = {
      resource: draft_resource,
      diffs: draft_resource.diffs,
      draft_values: draft_resource.draft_values,
      origin_values: draft_resource.origin_values
    }

    h.render("db/edit_requests/resource_diff_table", data)
  end

  def edit_draft_resource_path
    case draft_resource
    when DraftMultipleEpisode
      h.edit_db_work_draft_multiple_episode_path(draft_resource.work, draft_resource)
    when DraftWork
      h.edit_db_draft_work_path(draft_resource)
    when DraftEpisode
      h.edit_db_work_draft_episode_path(draft_resource.work, draft_resource)
    when DraftProgram
      h.edit_db_work_draft_program_path(draft_resource.work, draft_resource)
    when DraftItem
      h.edit_db_work_draft_item_path(draft_resource.work, draft_resource)
    when DraftPerson
      h.edit_db_draft_person_path(draft_resource)
    when DraftCast
      h.edit_db_person_draft_cast_path(draft_resource.person, draft_resource)
    end
  end
end
