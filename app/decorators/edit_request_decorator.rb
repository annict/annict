class EditRequestDecorator < Draper::Decorator
  delegate_all

  def edit_path
    case object.kind
    when "work"
      h.edit_db_works_edit_request_path(object)
    when "episodes"
      h.edit_db_work_episodes_edit_request_path(object.trackable, object)
    when "episode"
      h.edit_db_work_episode_edit_request_path(object.trackable, object.resource, object)
    end
  end

  def update_path
    case object.kind
    when "work"
      h.db_works_edit_request_path(object)
    end
  end

  def status_label
    if status.opened?
      h.content_tag :span, "オープン", class: "label label-success"
    elsif status.merged?
      h.content_tag :span, "反映済み", class: "label label-info"
    elsif status.closed?
      h.content_tag :span, "クローズ", class: "label label-danger"
    end
  end
end
