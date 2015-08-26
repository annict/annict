describe '作品別チャンネル設定ページ' do
  let(:user) { create(:registered_user) }
  let(:episode) { create(:episode) }
  let!(:program) { create(:program, episode: episode) }

  before do
    Status.skip_callback(:create, :after, :finish_tips)
    user.statuses.create(work: episode.work, kind: :watching)
    login_as(user, scope: :user)

    visit '/channel/works'
  end

  it 'ページが表示されること' do
    expect(page).to have_content(episode.work.title)
  end
end
