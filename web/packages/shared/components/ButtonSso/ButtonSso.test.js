/**
 * Copyright 2020 Gravitational, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import React from 'react';
import ButtonSso, { TypeEnum } from '.';
import { render } from 'design/utils/testing';

test.each`
  ssoType               | expectedIcon
  ${'default type'}     | ${'icon-openid'}
  ${TypeEnum.MICROSOFT} | ${'icon-windows'}
  ${TypeEnum.GITHUB}    | ${'icon-github'}
  ${TypeEnum.BITBUCKET} | ${'icon-bitbucket'}
  ${TypeEnum.GOOGLE}    | ${'icon-google-plus'}
`('rendering of $ssoType', ({ ssoType, expectedIcon }) => {
  const { getByTestId, getByText } = render(
    <ButtonSso ssoType={ssoType} title="hello" />
  );

  expect(getByTestId('icon')).toHaveClass(expectedIcon);
  expect(getByText(/hello/i)).toBeInTheDocument();
});
