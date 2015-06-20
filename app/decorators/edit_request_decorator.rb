class EditRequestDecorator < ApplicationDecorator
  def status_label
    if opened?
      h.content_tag :span, "オープン", class: "label label-success"
    elsif published?
      h.content_tag :span, "公開済み", class: "label label-info"
    elsif closed?
      h.content_tag :span, "クローズ", class: "label label-danger"
    end
  end

  def resource_diff_table
    data = {
      draft_resource: draft_resource,
      diffs: diffs,
      draft_values: draft_values,
      origin_values: origin_values
    }

    h.render("db/application/resource_diff_table", data)
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
    end
  end
end
