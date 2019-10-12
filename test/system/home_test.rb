# frozen_string_literal: true

require "application_system_test_case"

class HomeTest < ApplicationSystemTestCase
  setup do
    @work = create(:work, :with_current_season)
  end

  test "should display the hero words" do
    visit "/"
    assert_text "The platform for anime addicts."
  end
end

# describe "Top page" do
#   context "user does not sign in" do
#     let!(:work) { create(:work, :with_current_season) }
#
#     before do
#       visit "/"
#     end
#
#     it "displays the hero words" do
#       expect(page).to have_content("The platform for anime addicts.")
#     end
#   end
#
#   context "user signs in" do
#     let(:user) { create(:registered_user) }
#
#     before do
#       login_as(user, scope: :user)
#     end
#
#     describe "activity" do
#       context "no activities" do
#         before do
#           visit "/"
#         end
#
#         it "displays no activities message" do
#           expect(page).to have_content("アクティビティはありません")
#         end
#       end
#
#       context "has activities" do
#         let!(:activity) { create(:create_episode_record_activity, user: user) }
#
#         before do
#           visit "/"
#         end
#
#         it "displays activities" do
#           content = "#{user.profile.name}が#{activity.work.title}#{activity.episode.number}を見ました"
#           expect(page).to have_content(content)
#         end
#       end
#     end
#   end
# end
