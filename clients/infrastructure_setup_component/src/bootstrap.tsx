import { mountRemote } from '../../shared/runtime/mountRemote'
import { StandaloneNotice } from '../../shared/runtime/StandaloneNotice'

mountRemote(
  'infrastructure-setup-root',
  <StandaloneNotice title='Prompt Infrastructure Setup Component' />,
)
