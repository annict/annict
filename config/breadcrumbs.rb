# frozen_string_literal: true

crumb :root do
  link t("noun.home"), root_path
end

crumb :userland_root do
  link "Userland", userland_root_path
  parent :root
end

crumb :userland_new_project do
  link t("resources.userland_project.new"), new_userland_project_path
  parent :userland_root
end

crumb :userland_project_detail do |project|
  link project.name, userland_project_path(project)
  parent :userland_root
end

crumb :userland_edit_project do |project|
  link t("resources.userland_project.edit"), edit_userland_project_path(project)
  parent :userland_project_detail, project
end
