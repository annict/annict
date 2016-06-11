describe "放送予定ページ" do
  let!(:status_tip) { create(:status_tip) }
  let(:user) { create(:registered_user) }
  let(:work) { create(:work, :with_item) }
  let(:episode) { create(:episode, work: work) }
  let(:channel) { create(:channel) }

  before do
    login_as(user, scope: :user)

    channel.programs.create(episode: episode, work: work, started_at: Time.current)
    user.channel_works.create(work: work, channel: channel)
    user.statuses.create(work: work, kind: :watching)

    visit "/programs"
  end

  it "登録しているチャンネルで放送されるエピソードが表示されること", js: true do
    expect(find(".a-content .a-programs")).to have_content(episode.title)
  end
end
