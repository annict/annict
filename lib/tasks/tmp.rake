namespace :tmp do
  task create_db_activity: :environment do
    EditRequestComment.find_each do |erc|
      puts "id: #{erc.id}"
      DbActivity.where(
        user: erc.user,
        recipient: erc.edit_request,
        trackable: erc,
        action: "edit_request_comments.create"
      ).first_or_create
    end
  end
end
